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
// If inactivity_timeout is set, also kills the agent if it produces no output
// for that duration.
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

	inactivityKilled := r.readOutput(tctx, cancel, pr)
	pr.Close()

	cmd.Wait()

	if inactivityKilled || tctx.Err() == context.DeadlineExceeded {
		r.consecutiveTimeouts++
		logger.Error("agent timed out (consecutive: %d)", r.consecutiveTimeouts)
		return Result{TimedOut: true}
	}

	r.consecutiveTimeouts = 0
	return Result{}
}

// readOutput streams agent output to the logger. If inactivity_timeout is
// configured and no output arrives for that duration, it cancels the agent and
// returns true. Returns false on normal completion.
func (r *Runner) readOutput(ctx context.Context, cancel context.CancelFunc, pr *os.File) bool {
	inactivity := r.cfg.InactivityTimeout
	if inactivity == 0 {
		scanner := bufio.NewScanner(pr)
		for scanner.Scan() {
			logger.AgentLine(scanner.Text())
		}
		return false
	}

	lines := make(chan string)
	go func() {
		defer close(lines)
		scanner := bufio.NewScanner(pr)
		for scanner.Scan() {
			lines <- scanner.Text()
		}
	}()

	timer := time.NewTimer(inactivity)
	defer timer.Stop()

	for {
		select {
		case line, ok := <-lines:
			if !ok {
				return false
			}
			logger.AgentLine(line)
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			timer.Reset(inactivity)
		case <-timer.C:
			logger.Error("agent produced no output for %v — killing", inactivity)
			cancel()
			for range lines {} // drain so scanner goroutine can exit
			return true
		case <-ctx.Done():
			for range lines {}
			return false
		}
	}
}

func (r *Runner) ConsecutiveTimeouts() int {
	return r.consecutiveTimeouts
}
