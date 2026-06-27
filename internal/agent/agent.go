package agent

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/verity-bdd/verity-loop/internal/config"
	"github.com/verity-bdd/verity-loop/internal/logger"
)

// Runner executes the agent subprocess and tracks consecutive timeouts.
type Runner struct {
	cfg                 *config.Agent
	consecutiveTimeouts int
}

func New(cfg *config.Agent) *Runner {
	return &Runner{cfg: cfg}
}

type Result struct {
	TimedOut bool
	Err      error
}

// Run invokes the agent as: <command> <args...> <prompt>.
// Streams stdout+stderr to logger line by line. Enforces agent.timeout.
func (r *Runner) Run(ctx context.Context, prompt string) Result {
	timeout := r.cfg.Timeout
	if timeout == 0 {
		timeout = 10 * time.Minute
	}

	tctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	args := make([]string, len(r.cfg.Args)+1)
	copy(args, r.cfg.Args)
	args[len(r.cfg.Args)] = prompt

	cmd := exec.CommandContext(tctx, r.cfg.Command, args...)

	pr, pw, err := os.Pipe()
	if err != nil {
		return Result{Err: fmt.Errorf("creating pipe: %w", err)}
	}
	cmd.Stdout = pw
	cmd.Stderr = pw

	logger.Agent("running %s (timeout: %v)", r.cfg.Command, timeout)

	if err := cmd.Start(); err != nil {
		pw.Close()
		pr.Close()
		return Result{Err: fmt.Errorf("starting agent: %w", err)}
	}
	pw.Close()

	scanner := bufio.NewScanner(pr)
	for scanner.Scan() {
		logger.AgentLine(scanner.Text())
	}
	pr.Close()

	waitErr := cmd.Wait()

	if tctx.Err() == context.DeadlineExceeded {
		r.consecutiveTimeouts++
		logger.Error("agent timed out (consecutive: %d)", r.consecutiveTimeouts)
		return Result{TimedOut: true}
	}

	r.consecutiveTimeouts = 0

	if waitErr != nil {
		logger.Agent("warning: agent exited with error: %v", waitErr)
	}

	return Result{}
}

func (r *Runner) ConsecutiveTimeouts() int {
	return r.consecutiveTimeouts
}
