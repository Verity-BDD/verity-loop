package harness

import (
	"context"

	"github.com/verity-bdd/verity-loop/internal/agent"
	"github.com/verity-bdd/verity-loop/internal/config"
	"github.com/verity-bdd/verity-loop/internal/lifecycle"
	"github.com/verity-bdd/verity-loop/internal/logger"
	"github.com/verity-bdd/verity-loop/internal/prompt"
	"github.com/verity-bdd/verity-loop/internal/snapshot"
	"github.com/verity-bdd/verity-loop/internal/testrunner"
)

// Run executes the full harness loop and returns exit code (0 = success, 1 = failure).
// configPath must be the absolute path to verity.yaml.
func Run(ctx context.Context, configPath string) int {
	cfg, err := config.Load(configPath)
	if err != nil {
		logger.Error("config: %v", err)
		return 1
	}

	if err := config.CheckPromptFile(cfg); err != nil {
		logger.Error("%v", err)
		return 1
	}

	// INIT phase
	baseline, err := snapshot.TakeMulti(cfg.Services)
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
	result := testrunner.Run(cfg.ConfigDir, cfg.TestCommand)
	logger.Test(result.Passed, "preliminary test: %s", testStatus(result.Passed))
	if result.Passed {
		return 0
	}

	testOutput := testrunner.Truncate(result.Output, cfg.Context.MaxTestOutputLines)
	agentRunner := agent.New(&cfg.Agent)
	var rollbackDiffs []snapshot.ServiceDiff

	// Main loop
	for i := 1; i <= cfg.MaxIterations; i++ {
		select {
		case <-ctx.Done():
			return 1
		default:
		}

		logger.Init("iteration %d/%d", i, cfg.MaxIterations)

		preSnap, err := snapshot.TakeMulti(cfg.Services)
		if err != nil {
			logger.Error("pre-agent snapshot: %v", err)
			return 1
		}

		serviceDiffs := baseline.DiffAll(cfg.Context.MaxDiffLines)

		promptStr, err := prompt.Build(prompt.Params{
			Iteration:     i,
			PromptFile:    cfg.PromptFile,
			TestOutput:    testOutput,
			Services:      cfg.Services,
			ServiceDiffs:  serviceDiffs,
			RollbackDiffs: rollbackDiffs,
		})
		if err != nil {
			logger.Error("building prompt: %v", err)
			preSnap.Cleanup()
			return 1
		}
		rollbackDiffs = nil

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
			rollbackDiffs = preSnap.DiffAll(cfg.Context.MaxDiffLines)
			logger.Init("rolling back — service liveness failed after restart")
			if err := preSnap.RestoreAll(); err != nil {
				logger.Error("rollback warning: %v", err)
			}
			preSnap.Cleanup()
			continue
		}
		preSnap.Cleanup()

		// Check test result
		result = testrunner.Run(cfg.ConfigDir, cfg.TestCommand)
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
