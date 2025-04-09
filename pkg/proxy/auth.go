package proxy

import (
	"fmt"
	"log"
	"os"

	"github.com/aniqp/formal_assessment/pkg/config"
	"github.com/gliderlabs/ssh"
	gossh "golang.org/x/crypto/ssh"
)

type UserAuth struct {
	PublicKeyPath string
	Password      string
}

type AuthManager struct {
	Users map[string]UserAuth
}

func readPublicKey(keyPath string) (gossh.PublicKey, error) {
	keyData, err := os.ReadFile(keyPath)
	if err != nil {
		log.Printf("Failed to read public key for %s: %v", err)
		return nil, err
	}
	expectedKey, _, _, _, err := gossh.ParseAuthorizedKey(keyData)
	if err != nil {
		log.Printf("Failed to parse public key for %s: %v", err)
		return nil, err
	}
	return expectedKey, err
}

func readPrivateKey(path string) (ssh.Signer, error) {
	keyData, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read host key: %w", err)
	}

	signer, err := gossh.ParsePrivateKey(keyData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse host key: %w", err)
	}
	return signer, nil
}

func NewAuthManager(cfg *config.Config) *AuthManager {
	users := make(map[string]UserAuth)
	for _, u := range cfg.Proxy.AllowedUsers {
		users[u.Username] = UserAuth{
			PublicKeyPath: u.PublicKeyPath,
			Password:      u.Password,
		}
	}
	return &AuthManager{Users: users}
}

func (am *AuthManager) PublicKeyHandler(ctx ssh.Context, key ssh.PublicKey) bool {
	username := ctx.User()
	auth, exists := am.Users[username]
	if !exists {
		log.Printf("Public key auth failed: unknown user %s", username)
		return false
	}
	expectedKey, err := readPublicKey(auth.PublicKeyPath)
	if err != nil {
		log.Printf("Public key auth failed: %v", err)
		return false
	}

	return ssh.KeysEqual(key, expectedKey)
}

func (am *AuthManager) PasswordHandler(ctx ssh.Context, password string) bool {
	username := ctx.User()
	auth, exists := am.Users[username]
	if !exists {
		log.Printf("Password auth failed: Unknown user %s", username)
		return false
	}
	return auth.Password == password
}

func AddAuthMethods(keyPath string, password string) ([]gossh.AuthMethod, error) {
	var authMethods []gossh.AuthMethod
	if keyPath != "" {
		signer, err := readPrivateKey(keyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load sign key: %v", err)
		}
		authMethods = append(authMethods, gossh.PublicKeys(signer))
	}
	if password != "" {
		authMethods = append(authMethods, gossh.Password(password))
	}
	return authMethods, nil
}
