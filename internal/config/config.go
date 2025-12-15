package config

// DeploymentMode represents the deployment mode (swarm or compose)
type DeploymentMode string

const (
	// ModeSwarm uses Docker Swarm for deployment
	ModeSwarm DeploymentMode = "swarm"
	// ModeCompose uses docker compose for deployment
	ModeCompose DeploymentMode = "compose"
)

// Config represents the swarm.yaml configuration
type Config struct {
	Stack       string                `yaml:"stack"`
	Mode        DeploymentMode        `yaml:"mode"`
	SSH         SSHConfig             `yaml:"ssh"`
	Registry    Registry              `yaml:"registry"`
	Secrets     []string              `yaml:"secrets"`
	Accessories []string              `yaml:"accessories"`
	ComposeFile string                `yaml:"compose_file"`
	Nodes       map[string]NodeConfig `yaml:"nodes"`
}

// NodeConfig holds SSH settings for a specific node
type NodeConfig struct {
	User string `yaml:"user"`
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
		Mode: ModeSwarm,
		SSH: SSHConfig{
			Port: 22,
		},
		ComposeFile: "docker-compose.yaml",
	}
}
