# Spec: config-flag

## Purpose

Defines how `verity-loop run` accepts an optional `--config` flag to specify the location of `verity.yaml`, and how relative paths inside that file are resolved.

## Requirements

### Requirement: --config flag specifies verity.yaml location
The `verity-loop run` command SHALL accept an optional `--config <path>` flag. When provided, the harness loads the config file from that path. When absent, the harness looks for `verity.yaml` in the current working directory (existing behaviour).

#### Scenario: Run with explicit config path
- **WHEN** the user invokes `verity-loop run --config /projects/my-tests/verity.yaml`
- **THEN** the harness loads config from `/projects/my-tests/verity.yaml`

#### Scenario: Run without --config uses CWD
- **WHEN** the user invokes `verity-loop run` with no `--config` flag
- **THEN** the harness loads config from `./verity.yaml` in the current working directory

#### Scenario: --config with relative path
- **WHEN** the user invokes `verity-loop run --config ../other-project/verity.yaml`
- **THEN** the harness resolves the path relative to CWD and loads config from the resulting absolute path

#### Scenario: Non-existent config path fails with clear error
- **WHEN** the path supplied via `--config` does not exist
- **THEN** the harness exits with a clear error message identifying the missing file

### Requirement: Relative paths in config resolve from config file directory
All relative paths declared inside `verity.yaml` (including `work_dir`, `prompt_file`) SHALL be resolved relative to the directory that contains the `verity.yaml` file, not relative to the process CWD at invocation time.

#### Scenario: prompt_file resolved from config directory
- **WHEN** `verity.yaml` is at `/projects/tests/verity.yaml` and declares `prompt_file: ./PROMPT.md`
- **THEN** the harness reads `/projects/tests/PROMPT.md`

#### Scenario: work_dir resolved from config directory
- **WHEN** `verity.yaml` is at `/projects/tests/verity.yaml` and a service declares `work_dir: ../svc-a`
- **THEN** the harness uses `/projects/svc-a` as that service's effective `work_dir`

### Requirement: Agent and test command run from config file directory
When the harness launches the agent subprocess or executes the test command, the working directory SHALL be the directory containing `verity.yaml`.

#### Scenario: Agent CWD is config directory
- **WHEN** `verity.yaml` is at `/projects/tests/verity.yaml`
- **THEN** the agent subprocess is launched with CWD set to `/projects/tests/`

#### Scenario: Test command CWD is config directory
- **WHEN** `verity.yaml` is at `/projects/tests/verity.yaml`
- **THEN** `test_command` is executed with CWD set to `/projects/tests/`
