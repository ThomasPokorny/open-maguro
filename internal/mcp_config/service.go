package mcp_config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
)

var ErrServerNotFound = errors.New("MCP server not found")

// MCPConfigFile represents the mcp.json structure.
type MCPConfigFile struct {
	MCPServers map[string]MCPServer `json:"mcpServers"`
}

// MCPServer represents a single MCP server entry.
type MCPServer struct {
	Command string            `json:"command"`
	Args    []string          `json:"args"`
	Env     map[string]string `json:"env,omitempty"`
}

type AddServerRequest struct {
	Name    string            `json:"name"    validate:"required,min=1,max=100"`
	Command string            `json:"command" validate:"required"`
	Args    []string          `json:"args"    validate:"required"`
	Env     map[string]string `json:"env"`
}

type ServerResponse struct {
	Name    string            `json:"name"`
	Command string            `json:"command"`
	Args    []string          `json:"args"`
	Env     map[string]string `json:"env,omitempty"`
}

type Service struct {
	configPath string
	mu         sync.Mutex
}

func NewService(configPath string) *Service {
	return &Service{configPath: configPath}
}

func (s *Service) List() ([]ServerResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	cfg, err := s.read()
	if err != nil {
		return nil, err
	}

	servers := make([]ServerResponse, 0, len(cfg.MCPServers))
	for name, server := range cfg.MCPServers {
		servers = append(servers, ServerResponse{
			Name:    name,
			Command: server.Command,
			Args:    server.Args,
			Env:     server.Env,
		})
	}
	return servers, nil
}

func (s *Service) Add(req AddServerRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	cfg, err := s.read()
	if err != nil {
		return err
	}

	cfg.MCPServers[req.Name] = MCPServer{
		Command: req.Command,
		Args:    req.Args,
		Env:     req.Env,
	}

	return s.write(cfg)
}

func (s *Service) Remove(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	cfg, err := s.read()
	if err != nil {
		return err
	}

	if _, ok := cfg.MCPServers[name]; !ok {
		return ErrServerNotFound
	}

	delete(cfg.MCPServers, name)
	return s.write(cfg)
}

func (s *Service) read() (*MCPConfigFile, error) {
	data, err := os.ReadFile(s.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &MCPConfigFile{MCPServers: make(map[string]MCPServer)}, nil
		}
		return nil, fmt.Errorf("read MCP config: %w", err)
	}

	var cfg MCPConfigFile
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse MCP config: %w", err)
	}

	if cfg.MCPServers == nil {
		cfg.MCPServers = make(map[string]MCPServer)
	}

	return &cfg, nil
}

func (s *Service) write(cfg *MCPConfigFile) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal MCP config: %w", err)
	}

	data = append(data, '\n')

	if err := os.WriteFile(s.configPath, data, 0644); err != nil {
		return fmt.Errorf("write MCP config: %w", err)
	}

	return nil
}
