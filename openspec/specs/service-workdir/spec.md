# Spec: service-workdir

## Purpose

Defines how each service in `verity.yaml` may declare a `work_dir` field, how that directory is resolved, and how it affects lifecycle commands, git operations, agent prompts, and diff output.

## Requirements

### Requirement: Service declares work_dir
Each service in `verity.yaml` MAY declare a `work_dir` field (string). When present, it specifies the root directory of that service's code. When absent, the service's effective `work_dir` SHALL default to the directory containing `verity.yaml`.

#### Scenario: Service with explicit work_dir
- **WHEN** a service entry includes `work_dir: /projects/svc-a`
- **THEN** the harness uses `/projects/svc-a` as the working directory for that service's lifecycle commands and git operations

#### Scenario: Service without work_dir defaults to config directory
- **WHEN** a service entry omits `work_dir`
- **THEN** the harness uses the directory containing `verity.yaml` as the effective `work_dir` for that service

#### Scenario: Relative work_dir resolved from config directory
- **WHEN** a service entry includes `work_dir: ../svc-a`
- **THEN** the harness resolves the path relative to the directory containing `verity.yaml` and uses the resulting absolute path

### Requirement: Lifecycle commands run from service work_dir
The `start`, `stop`, and `restart` shell commands for a service SHALL be executed with the service's resolved `work_dir` as the process working directory.

#### Scenario: Start command runs in work_dir
- **WHEN** the harness starts a service with `work_dir: /projects/svc-a` and `start: make run`
- **THEN** `make run` is executed with CWD set to `/projects/svc-a`

#### Scenario: Restart command runs in work_dir
- **WHEN** the harness restarts a service after an agent iteration
- **THEN** the `restart` command runs with the service's resolved `work_dir` as CWD

### Requirement: Git operations scoped to service work_dir
The harness SHALL take git snapshots, compute diffs, and perform rollbacks independently in each service's resolved `work_dir`.

#### Scenario: Baseline snapshot taken per service
- **WHEN** the harness initialises before the first iteration
- **THEN** a git snapshot is taken in each service's `work_dir`

#### Scenario: Pre-agent snapshot taken per service
- **WHEN** the harness begins each iteration before running the agent
- **THEN** a git snapshot is taken in each service's `work_dir`

#### Scenario: Rollback restores all service work_dirs
- **WHEN** any service fails its liveness check after a restart
- **THEN** the harness rolls back the git working tree to the pre-agent snapshot state in every service's `work_dir`

#### Scenario: Service work_dir without git repository fails at startup
- **WHEN** a service's resolved `work_dir` is not inside a git repository
- **THEN** the harness exits with an error during the INIT phase before any service is started

### Requirement: Agent prompt includes service locations
The agent prompt SHALL include a "Services" section listing each service's name and resolved `work_dir`, inserted after the user prompt content (iteration 1) or after the test output (iterations 2+).

#### Scenario: Services section present in iteration 1 prompt
- **WHEN** the harness builds the iteration 1 prompt
- **THEN** the prompt contains a `--- Services ---` section with one line per service: `<name>: <resolved_work_dir>`

#### Scenario: Services section present in later iteration prompts
- **WHEN** the harness builds any prompt for iteration 2 or later
- **THEN** the prompt contains the same `--- Services ---` section

### Requirement: Per-service diff sections in prompt
When the harness includes a diff in the prompt (iterations 2+), it SHALL produce one labelled section per service that has changes. Services with empty diffs are omitted.

#### Scenario: Diff section labelled with service name and path
- **WHEN** svc-a has changes and svc-b has no changes after the previous iteration
- **THEN** the prompt contains `--- Your changes in svc-a (/projects/svc-a) ---` followed by the diff, and no section for svc-b

#### Scenario: Multiple services with changes produce multiple sections
- **WHEN** both svc-a and svc-b have changes after the previous iteration
- **THEN** the prompt contains separate labelled diff sections for each service in declaration order
