package harness

import (
	"context"
	"path/filepath"

	"github.com/nikchursin/verity-harness/internal/agent"
	"github.com/nikchursin/verity-harness/internal/config"
	"github.com/nikchursin/verity-harness/internal/lifecycle"
	"github.com/nikchursin/verity-harness/internal/logger"
	"github.com/nikchursin/verity-harness/internal/prompt"
	"github.com/nikchursin/verity-harness/internal/snapshot"
	"github.com/nikchursin/verity-harness/internal/testrunner"
)

// Run executes the full harness loop and returns exit code (0 = success, 1 = failure).
// ctx should be cancelled on SIGINT/SIGTERM by the caller.
func Run(ctx context.Context, workDir string) int {
	cfg, err := config.Load(filepath.Join(workDir, "verity.yaml"))
	if err != nil {
		logger.Error("config: %v", err)
		return 1
	}

	if err := config.CheckPromptFile(cfg); err != nil {
		logger.Error("%v", err)
		return 1
	}

	// INIT phase
	baseline, err := snapshot.TakeSnapshot(workDir)
	if err != nil {
		logger.Error("baseline snapshot: %v", err)
		return 1
	}
	defer baseline.Cleanup()

	mgr := lifecycle.New(cfg.Services)
	defer mgr.Teardown()

	if err := mgr.StartAll(ctx); err != nil {
		logger.Error("starting services: %v", err)
		return 1
	}

	// Preliminary test — exit early if already passing
	result := testrunner.Run(workDir, cfg.TestCommand)
	logger.Test(result.Passed, "preliminary test: %s", testStatus(result.Passed))
	if result.Passed {
		return 0
	}

	testOutput := testrunner.Truncate(result.Output, cfg.Context.MaxTestOutputLines)
	agentRunner := agent.New(&cfg.Agent)
	var rollbackDiff string

	// Main loop
	for i := 1; i <= cfg.MaxIterations; i++ {
		select {
		case <-ctx.Done():
			return 1
		default:
		}

		logger.Init("iteration %d/%d", i, cfg.MaxIterations)

		preSnap, err := snapshot.TakeSnapshot(workDir)
		if err != nil {
			logger.Error("pre-agent snapshot: %v", err)
			return 1
		}

		baselineDiff, _ := snapshot.Diff(workDir, baseline)

		promptStr, err := prompt.Build(prompt.Params{
			Iteration:    i,
			PromptFile:   cfg.PromptFile,
			TestOutput:   testOutput,
			BaselineDiff: baselineDiff,
			RollbackDiff: rollbackDiff,
			MaxDiffLines: cfg.Context.MaxDiffLines,
		})
		if err != nil {
			logger.Error("building prompt: %v", err)
			preSnap.Cleanup()
			return 1
		}
		rollbackDiff = ""

		// Run agent
		agentResult := agentRunner.Run(ctx, promptStr)
		if agentResult.TimedOut {
			preSnap.Cleanup()
			if agentRunner.ConsecutiveTimeouts() >= 3 {
				logger.Error("agent timed out 3 times in a row")
				return 1
			}
			continue
		}

		// Restart services and check liveness
		if ok := mgr.Restart(ctx); !ok {
			// Liveness failed — capture diff of what agent did, then rollback
			agentDiff, _ := snapshot.Diff(workDir, preSnap)
			logger.Init("rolling back — service liveness failed after restart")
			if err := snapshot.Restore(workDir, preSnap); err != nil {
				logger.Error("rollback warning: %v", err)
			}
			rollbackDiff = agentDiff
			preSnap.Cleanup()
			continue
		}
		preSnap.Cleanup()

		// Check test result
		result = testrunner.Run(workDir, cfg.TestCommand)
		testOutput = testrunner.Truncate(result.Output, cfg.Context.MaxTestOutputLines)
		logger.Test(result.Passed, "iteration %d: %s", i, testStatus(result.Passed))

		if result.Passed {
			return 0
		}
	}

	logger.Error("exhausted %d iterations — test still failing", cfg.MaxIterations)
	return 1
}

func testStatus(passed bool) string {
	if passed {
		return "PASS"
	}
	return "FAIL"
}
