package config

// Config represents the swarm.yaml configuration
type Config struct {
	Stack       string      `yaml:"stack"`
	SSH         SSHConfig   `yaml:"ssh"`
	Registry    Registry    `yaml:"registry"`
	Secrets     []string    `yaml:"secrets"`
	Accessories []string    `yaml:"accessories"`
	ComposeFile string      `yaml:"compose_file"`
}

// SSHConfig holds SSH connection settings
type SSHConfig struct {
	Host string `yaml:"host"`
	User string `yaml:"user"`
	Port int    `yaml:"port"`
	Key  string `yaml:"key"`
}

// Registry holds container registry settings
type Registry struct {
	URL      string `yaml:"url"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

// NewConfig returns a Config with default values
func NewConfig() *Config {
	return &Config{
		SSH: SSHConfig{
			Port: 22,
		},
		ComposeFile: "docker-compose.yaml",
	}
}
