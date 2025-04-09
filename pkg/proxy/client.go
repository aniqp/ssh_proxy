package proxy

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/gliderlabs/ssh"
	gossh "golang.org/x/crypto/ssh"
)

func (h *SessionHandler) connectUpstream(addr string, sshConfig *gossh.ClientConfig, s ssh.Session) error {
	client, err := gossh.Dial("tcp", addr, sshConfig)
	if err != nil {
		return fmt.Errorf("unable to connect to upstream server: %v", err)
	}
	defer client.Close()

	upstreamSession, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %v", err)
	}
	defer upstreamSession.Close()

	logFile, err := h.logger.createLogFile(s.User(), time.Now(), h.cfg.Proxy.LogDir)
	if err != nil {
		return fmt.Errorf("failed to create log file: %s", err)
	}

	logFilePath := logFile.Name()

	forwardIO(upstreamSession, s, logFile)

	ptyReq, winCh, isPty := s.Pty()
	if isPty {
		err := h.handleInteractiveSession(upstreamSession, ptyReq, winCh)
		if err != nil {
			logFile.Close()
			return fmt.Errorf("error handling interactive session: %w", err)
		}
	} else {
		command := s.RawCommand()
		_, err := logFile.WriteString(command + "\n")
		if err != nil {
			logFile.Close()
			return fmt.Errorf("error writing to log file: %v", err)
		}
		err = upstreamSession.Run(command)
		if err != nil {
			logFile.Close()
			return fmt.Errorf("failed to run command: %v", err)
		}
	}

	if err := logFile.Close(); err != nil {
		log.Printf("Failed to close log file: %v", err)
	}

	if h.cfg.LLM.Enabled {
		go func(path string, user string) {
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()

			if err := h.llmService.summarizeSession(ctx, path, user); err != nil {
				log.Printf("Error while summarizing session for %s: %v", user, err)
			}
		}(logFilePath, s.User())
	}

	return nil
}

func (h *SessionHandler) handleInteractiveSession(us *gossh.Session, ptyReq ssh.Pty, winCh <-chan ssh.Window) error {
	if err := us.RequestPty(ptyReq.Term, ptyReq.Window.Height, ptyReq.Window.Width, gossh.TerminalModes{}); err != nil {
		return fmt.Errorf("failed to request PTY: %v", err)
	}

	go func() {
		for win := range winCh {
			if err := us.WindowChange(win.Height, win.Width); err != nil {
				log.Printf("failed to change window size: %v", err)
			}
		}
	}()

	if err := us.Shell(); err != nil {
		return fmt.Errorf("failed to start shell: %v", err)
	}

	err := us.Wait()
	if err != nil {
		return fmt.Errorf("exited terminal with error: %w", err)
	}
	return nil
}

// forwarding stdin and stdout of upstream session to proxy session and log file
func forwardIO(us *gossh.Session, s ssh.Session, f *os.File) {
	stdout, _ := us.StdoutPipe()
	stderr, _ := us.StderrPipe()
	stdin, _ := us.StdinPipe()

	go io.Copy(io.MultiWriter(stdin, f), s)
	go io.Copy(s, stdout)
	go io.Copy(s.Stderr(), stderr)
}
