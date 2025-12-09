package config

import (
	"flag"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type User struct {
	Username      string `yaml:"username"`
	PublicKeyPath string `yaml:"public_key_path"`
	Password      string `yaml:"password"`
}

type ProxyConfig struct {
	ListenAddress string `yaml:"listen_address"`
	LogDir        string `yaml:"log_dir"`
	HostKey       string `yaml:"host_key"`
	AllowedUsers  []User `yaml:"allowed_users"`
}

type UpstreamConfig struct {
	Host           string `yaml:"host"`
	Port           int    `yaml:"port"`
	Username       string `yaml:"username"`
	PrivateKeyPath string `yaml:"client_private_key_path"`
	HostKeyPath    string `yaml:"host_key_path"`
	Password       string `yaml:"password"`
}

type LLMConfig struct {
	APIKey  string `yaml:"api_key"`
	Prompt  string `yaml:"prompt"`
	Enabled bool
}

type Config struct {
	Proxy    ProxyConfig    `yaml:"proxy"`
	Upstream UpstreamConfig `yaml:"upstream"`
	LLM      LLMConfig      `yaml:"llm"`
}

func LoadConfig() (*Config, error) {
	configPath := flag.String("config", "", "Path to the configuration file")
	flag.Parse()

	if *configPath == "" {
		return nil, fmt.Errorf("please specify a configuration file")
	}

	f, err := os.ReadFile(*configPath)
	if err != nil {
		return nil, err
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(f, cfg); err != nil {
		return nil, err
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	if c.Proxy.ListenAddress == "" {
		return fmt.Errorf("proxy listen_address is required")
	}
	if c.Proxy.LogDir == "" {
		return fmt.Errorf("proxy log_dir is required")
	}
	if c.Proxy.HostKey == "" {
		return fmt.Errorf("proxy host_key is required")
	}
	if len(c.Proxy.AllowedUsers) == 0 {
		return fmt.Errorf("at least one allowed user is required")
	}

	for i, user := range c.Proxy.AllowedUsers {
		if user.Username == "" {
			return fmt.Errorf("user [%d] username is required", i)
		}

		if user.PublicKeyPath == "" && user.Password == "" {
			return fmt.Errorf("user [%d] public key path or password is required", i)
		}
	}

	if c.Upstream.Host == "" {
		return fmt.Errorf("upstream host is required")
	}

	if c.Upstream.Port == 0 {
		return fmt.Errorf("please specify the port for the upstream server")
	}

	if c.Upstream.Username == "" {
		return fmt.Errorf("proxy's upstream username is required")
	}

	if c.Upstream.HostKeyPath == "" {
		return fmt.Errorf("upstream host_key_path is required")
	}

	if c.Upstream.PrivateKeyPath == "" && c.Upstream.Password == "" {
		return fmt.Errorf("either upstream private_key_path or upstream password must be provided")
	}

	return nil
}
