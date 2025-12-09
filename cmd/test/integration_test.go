package test

import (
	"bytes"
	"os/exec"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	proxyAddr    = "localhost"
	proxyPort    = "2022"
	upstreamAddr = "proxyserver"
	testUser     = "user2"
	privateKey   = "../../keys/client/priv/user2_ecdsa"
)

func runSSHCommand(cmd string) (string, error) {
	sshCmd := exec.Command("ssh",
		"-i", privateKey,
		"-p", proxyPort,
		testUser+"@"+proxyAddr,
		cmd,
	)
	var out bytes.Buffer
	sshCmd.Stdout = &out
	sshCmd.Stderr = &out
	err := sshCmd.Run()
	return out.String(), err
}

func startInteractiveSession(t *testing.T) {
	sshCmd := exec.Command("ssh",
		"-i", privateKey,
		"-p", proxyPort,
		"-tt",
		testUser+"@"+proxyAddr,
	)

	var out bytes.Buffer
	sshCmd.Stdout = &out
	sshCmd.Stderr = &out
	inPipe, _ := sshCmd.StdinPipe()

	err := sshCmd.Start()
	assert.NoError(t, err, "Failed to start interactive session")

	_, err = inPipe.Write([]byte("echo interactive_test\n"))
	assert.NoError(t, err, "Failed to write to interactive session")

	time.Sleep(2 * time.Second)
	assert.Contains(t, out.String(), "interactive_test", "Interactive session did not return expected output")

	err = sshCmd.Process.Kill()
	assert.NoError(t, err, "Failed to kill interactive session")
}

func TestSingleClientMultipleCommands(t *testing.T) {
	time.Sleep(2 * time.Second)

	output, err := runSSHCommand("whoami && hostname && date")
	assert.NoError(t, err, "SSH command execution failed")

	assert.Contains(t, output, upstreamAddr, "Username should be returned")
	assert.Contains(t, output, "UTC", "Timestamp should be in UTC format")
}

func TestMultipleClientSessions(t *testing.T) {
	cmds := map[string]string{
		"hostname":  "openssh-server",
		"whoami":    upstreamAddr,
		"echo test": "test",
	}

	var wg sync.WaitGroup
	results := make(map[string]string)
	var mu sync.Mutex

	for cmd, expectedOutput := range cmds {
		wg.Add(1)
		go func(c, expected string) {
			defer wg.Done()
			out, err := runSSHCommand(c)
			assert.NoError(t, err, "SSH command execution failed")

			mu.Lock()
			results[c] = out
			mu.Unlock()

			assert.Contains(t, out, expected, "Unexpected output for command: "+c)
		}(cmd, expectedOutput)
	}
	wg.Wait()
}

func TestSingleClientInteractiveSessions(t *testing.T) {
	startInteractiveSession(t)
}

func TestMultipleClientsInteractiveSessions(t *testing.T) {
	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(clientID int) {
			defer wg.Done()
			startInteractiveSession(t)
		}(i)
	}
	wg.Wait()
}
