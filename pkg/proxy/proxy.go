package proxy

import (
	"fmt"

	"github.com/aniqp/formal_assessment/pkg/config"
	"github.com/gliderlabs/ssh"
)

type ProxyServer struct {
	cfg       *config.Config
	SSHServer *ssh.Server
}

func NewServer(cfg *config.Config) (*ProxyServer, error) {
	am := NewAuthManager(cfg)
	signer, err := readPrivateKey(cfg.Proxy.HostKey)
	if err != nil {
		return nil, fmt.Errorf("failed to load host key: %w", err)
	}

	server := &ProxyServer{
		cfg: cfg,
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
