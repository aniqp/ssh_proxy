package proxy

import (
	"fmt"
	"github.com/aniqp/formal_assessment/pkg/config"
	"github.com/creack/pty"
	"github.com/gliderlabs/ssh"
	"os"
)

type Server struct {
	Cfg       *config.Config
	SSHServer *ssh.Server
}

func setWinsize(f *os.File, width, height int) error {
	return pty.Setsize(f, &pty.Winsize{
		Cols: uint16(width),
		Rows: uint16(height),
	})
}

func NewServer(cfg *config.Config) (*Server, error) {
	am := NewAuthManager(cfg)
	signer, err := readPrivateKey(cfg.Proxy.HostKey)
	if err != nil {
		return nil, fmt.Errorf("failed to load host key: %w", err)
	}

	server := &Server{
		Cfg: cfg,
		SSHServer: &ssh.Server{
			Addr:             cfg.Proxy.ListenAddress,
			Handler:          NewSessionHandler(cfg),
			PublicKeyHandler: am.PublicKeyHandler,
			PasswordHandler:  am.PasswordHandler,
		},
	}
	server.SSHServer.AddHostKey(signer)
	return server, nil
}
