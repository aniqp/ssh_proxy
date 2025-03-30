package config

import (
	"flag"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
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
	Host                string `yaml:"host"`
	Port                int    `yaml:"port"`
	Username            string `yaml:"username"`
	PrivateKeyPath      string `yaml:"private_key_path"`
	UpstreamHostKeyPath string `yaml:"upstream_host_key_path"`
	Password            string `yaml:"password"`
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
		return fmt.Errorf("proxy_listen_address is required")
	}
	if c.Upstream.Host == "" {
		return fmt.Errorf("host is required")
	}
	if c.Upstream.Port == 0 {
		return fmt.Errorf("port is required")
	}
	if len(c.Proxy.AllowedUsers) == 0 {
		return fmt.Errorf("at least one allowed user is required")
	}
	return nil
}
