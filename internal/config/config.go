package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Agent         Agent     `yaml:"agent"`
	TestCommand   string    `yaml:"test_command"`
	PromptFile    string    `yaml:"prompt_file"`
	MaxIterations int       `yaml:"max_iterations"`
	Services      []Service `yaml:"services"`
	Context       Context   `yaml:"context"`
	ConfigDir     string    `yaml:"-"`
}

type Agent struct {
	Command           string        `yaml:"command"`
	Args              []string      `yaml:"args"`
	Timeout           time.Duration `yaml:"timeout"`
	InactivityTimeout time.Duration `yaml:"inactivity_timeout"`
}

type Service struct {
	Name     string            `yaml:"name"`
	Start    string            `yaml:"start"`
	Stop     string            `yaml:"stop"`
	Restart  string            `yaml:"restart"`
	WorkDir  string            `yaml:"work_dir"`
	Env      map[string]string `yaml:"env"`
	Liveness Liveness          `yaml:"liveness"`
}

type Liveness struct {
	URL      string        `yaml:"url"`
	Interval time.Duration `yaml:"interval"`
	Timeout  time.Duration `yaml:"timeout"`
}

type Context struct {
	MaxDiffLines       int `yaml:"max_diff_lines"`
	MaxTestOutputLines int `yaml:"max_test_output_lines"`
}

func Load(path string) (*Config, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolving config path: %w", err)
	}
	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("cannot read %s: %w", path, err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("invalid YAML in %s: %w", path, err)
	}
	cfg.ConfigDir = filepath.Dir(absPath)
	applyDefaults(&cfg)
	resolveRelativePaths(&cfg)
	if err := validate(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func resolveRelativePaths(cfg *Config) {
	if cfg.PromptFile != "" && !filepath.IsAbs(cfg.PromptFile) {
		cfg.PromptFile = filepath.Join(cfg.ConfigDir, cfg.PromptFile)
	}
	for i := range cfg.Services {
		if cfg.Services[i].WorkDir == "" {
			cfg.Services[i].WorkDir = cfg.ConfigDir
		} else if !filepath.IsAbs(cfg.Services[i].WorkDir) {
			cfg.Services[i].WorkDir = filepath.Join(cfg.ConfigDir, cfg.Services[i].WorkDir)
		}
	}
}

func applyDefaults(cfg *Config) {
	if cfg.MaxIterations == 0 {
		cfg.MaxIterations = 10
	}
	if cfg.Agent.Timeout == 0 {
		cfg.Agent.Timeout = 10 * time.Minute
	}
	if cfg.Context.MaxDiffLines == 0 {
		cfg.Context.MaxDiffLines = 200
	}
	if cfg.Context.MaxTestOutputLines == 0 {
		cfg.Context.MaxTestOutputLines = 100
	}
	for i := range cfg.Services {
		if cfg.Services[i].Liveness.Interval == 0 {
			cfg.Services[i].Liveness.Interval = time.Second
		}
	}
}

func validate(cfg *Config) error {
	if cfg.Agent.Command == "" {
		return fmt.Errorf("missing required field: agent.command")
	}
	if cfg.TestCommand == "" {
		return fmt.Errorf("missing required field: test_command")
	}
	if cfg.PromptFile == "" {
		return fmt.Errorf("missing required field: prompt_file")
	}
	if len(cfg.Services) == 0 {
		return fmt.Errorf("missing required field: services (must have at least one)")
	}
	return nil
}

func CheckPromptFile(cfg *Config) error {
	if _, err := os.Stat(cfg.PromptFile); os.IsNotExist(err) {
		return fmt.Errorf("prompt_file not found: %s", cfg.PromptFile)
	}
	return nil
}
