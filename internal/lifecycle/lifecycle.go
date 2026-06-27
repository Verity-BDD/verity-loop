package lifecycle

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/verity-bdd/verity-loop/internal/config"
	"github.com/verity-bdd/verity-loop/internal/logger"
)

type Manager struct {
	services []config.Service
	running  []*exec.Cmd
}

func New(services []config.Service) *Manager {
	return &Manager{
		services: services,
		running:  make([]*exec.Cmd, len(services)),
	}
}

// StartAll starts each service in order and waits for liveness before moving to the next.
func (m *Manager) StartAll(ctx context.Context) error {
	for i, svc := range m.services {
		logger.Init("starting %s", svc.Name)
		cmd, err := m.startBackground(svc)
		if err != nil {
			return fmt.Errorf("starting %s: %w", svc.Name, err)
		}
		m.running[i] = cmd

		if svc.Liveness.URL != "" {
			if ok := m.pollLiveness(ctx, svc); !ok {
				return fmt.Errorf("liveness timeout for %s", svc.Name)
			}
		}
	}
	return nil
}

// Restart runs restart commands for all services in order and polls liveness.
// Returns false if any liveness check fails.
func (m *Manager) Restart(ctx context.Context) bool {
	for _, svc := range m.services {
		logger.Restart("restarting %s", svc.Name)
		if svc.Restart != "" {
			if err := m.runSync(ctx, svc.Restart, svc.Env, svc.WorkDir); err != nil {
				logger.Error("restart command failed for %s: %v", svc.Name, err)
				return false
			}
		}
		if svc.Liveness.URL != "" {
			if ok := m.pollLiveness(ctx, svc); !ok {
				logger.Live(false, "%s liveness failed after restart", svc.Name)
				return false
			}
		}
	}
	return true
}

// Teardown stops all services in reverse order with a 10s total timeout.
func (m *Manager) Teardown() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for i := len(m.services) - 1; i >= 0; i-- {
		svc := m.services[i]
		logger.Stop("stopping %s", svc.Name)
		if svc.Stop != "" {
			if err := m.runSync(ctx, svc.Stop, svc.Env, svc.WorkDir); err != nil {
				logger.Error("stop command failed for %s: %v", svc.Name, err)
			}
		}
		if m.running[i] != nil && m.running[i].Process != nil {
			m.running[i].Process.Kill()
		}
	}
}

func (m *Manager) startBackground(svc config.Service) (*exec.Cmd, error) {
	cmd := exec.Command("sh", "-c", svc.Start)
	cmd.Dir = svc.WorkDir
	cmd.Env = buildEnv(svc.Env)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return cmd, nil
}

func (m *Manager) runSync(ctx context.Context, command string, env map[string]string, workDir string) error {
	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	cmd.Dir = workDir
	cmd.Env = buildEnv(env)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (m *Manager) pollLiveness(ctx context.Context, svc config.Service) bool {
	timeout := svc.Liveness.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	interval := svc.Liveness.Interval
	if interval == 0 {
		interval = time.Second
	}

	deadline := time.Now().Add(timeout)
	start := time.Now()
	throttle := &logger.LiveThrottle{}
	client := &http.Client{Timeout: 2 * time.Second}

	for {
		select {
		case <-ctx.Done():
			return false
		default:
		}

		if time.Now().After(deadline) {
			return false
		}

		resp, err := client.Get(svc.Liveness.URL)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				logger.Live(true, "%s is alive", svc.Name)
				return true
			}
		}

		throttle.Log(time.Since(start))

		select {
		case <-ctx.Done():
			return false
		case <-time.After(interval):
		}
	}
}

func buildEnv(extra map[string]string) []string {
	base := os.Environ()
	for k, v := range extra {
		base = append(base, k+"="+v)
	}
	return base
}
