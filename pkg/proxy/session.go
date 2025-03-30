package proxy

import (
	"errors"
	"github.com/aniqp/formal_assessment/pkg/config"
	"github.com/gliderlabs/ssh"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	gossh "golang.org/x/crypto/ssh"
	"log"
	"strconv"
)

type SessionHandler struct {
	cfg        *config.Config
	llmService *LLMService
	logger     *Logger
}

func NewSessionHandler(cfg *config.Config) ssh.Handler {
	client := openai.NewClient(
		option.WithAPIKey(cfg.LLM.APIKey),
	)
	logger := NewLogger(cfg)
	llmService := NewLLMService(cfg, client)
	llmService.validateLLMKey()

	handler := &SessionHandler{
		cfg:        cfg,
		llmService: llmService,
		logger:     logger,
	}
	return handler.Handle
}

func (h *SessionHandler) Handle(s ssh.Session) {
	authMethods, err := AddAuthMethods(h.cfg.Upstream.PrivateKeyPath, h.cfg.Upstream.Password)
	if err != nil {
		log.Fatalf("failed to add auth methods: %v", err)
	}

	upstreamKey, err := readPublicKey(h.cfg.Upstream.UpstreamHostKeyPath)
	if err != nil {
		log.Fatalf("failed to read upstream private key: %v", err)
	}
	sshConfig := &gossh.ClientConfig{
		User:            h.cfg.Upstream.Username,
		Auth:            authMethods,
		HostKeyCallback: gossh.FixedHostKey(upstreamKey),
	}

	addr := h.cfg.Upstream.Host + ":" + strconv.Itoa(h.cfg.Upstream.Port)
	if err := h.connectUpstream(addr, sshConfig, s); err != nil {
		var ce *ConnectionErr
		var ee *gossh.ExitError

		if errors.As(err, &ce) {
			log.Fatalf("failed to connect upstream: %v", err)
		} else if errors.As(err, &ee) {
			log.Printf("session exited with status: %d", ee.ExitStatus())
		} else {
			log.Printf("unexpected error in connecting upstream: %v", err)
		}
	}
}
