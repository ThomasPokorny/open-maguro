package config

type Config struct {
	Port          string   `env:"PORT"            envDefault:"8080"`
	DatabaseURL   string   `env:"DATABASE_URL,required"`
	LogLevel      string   `env:"LOG_LEVEL"       envDefault:"info"`
	MCPConfigPath string   `env:"MCP_CONFIG_PATH"`
	AllowedTools  []string `env:"ALLOWED_TOOLS"   envSeparator:"," envDefault:"Bash(curl*),Bash(npx*),WebSearch,WebFetch,mcp__*"`
	WorkspaceRoot string   `env:"WORKSPACE_ROOT"  envDefault:"~/.maguro/workspaces"`
}
