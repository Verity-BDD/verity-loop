# Verity Loop — Specification v0.4

## Концепция

CLI-инструмент для разработчика. Поднимает сервисы, принимает падающий acceptance-тест (Go), запускает LLM-агента для написания кода, итеративно проверяет тест пока не станет зелёным.

---

## Команда запуска

```
verity-loop run
```

Промпт, список сервисов и тестовая команда берутся из `verity.yaml`.

---

## Полный цикл

```
INIT:
  baseline_snapshot = git snapshot
  for each service (in order):
    execute service.start (with env)
    poll service.liveness until 200 (timeout: configurable)
    FAIL → teardown → exit 1

LOOP (max_iterations):
  pre_agent_snapshot = git snapshot

  run test_command → capture output
  if green → teardown → exit 0

  iteration 1:          prompt = user_prompt + test_output
  iteration 2+ normal:  prompt = test_output + diff(vs baseline_snapshot)
  iteration 2+ rollback: prompt = test_output
                               + "Previous attempt broke service restart:"
                               + diff(vs pre_agent_snapshot)

  execute agent with prompt (subprocess, no TTY, timeout: agent.timeout)
    stream agent stdout to log
    if timeout:
      consecutive_timeouts++
      if consecutive_timeouts >= 3 → teardown → exit 1
      else → skip restart, next iteration
    if success:
      consecutive_timeouts = 0

  for each service (in order):
    execute service.restart
    poll service.liveness until 200:
      OK   → continue
      FAIL → restore pre_agent_snapshot
             mark next iteration as rollback
             break

EXHAUSTED → teardown → exit 1
SIGINT/SIGTERM → teardown → exit 1

teardown():
  for each service (reverse order):
    execute service.stop
```

---

## Конфигурация (`verity.yaml`)

```yaml
agent:
  command: "opencode"
  args: ["run"]
  timeout: 10m              # default: 10m

max_iterations: 10

prompt_file: "./PROMPT.md"

test_command: "go test ./tests/ -run TestMyFeature -v"

context:
  max_diff_lines: 200
  max_test_output_lines: 100

services:
  - name: svc-a
    start: "make -C ./svc-a run"
    stop: "make -C ./svc-a stop"
    restart: "make -C ./svc-a restart"
    env:
      SVC_B_URL: "http://localhost:8081"
    liveness:
      url: "http://localhost:8080/health"
      timeout: 30s
      interval: 1s

  - name: svc-b
    start: "make -C ./svc-b run"
    stop: "make -C ./svc-b stop"
    restart: "make -C ./svc-b restart"
    liveness:
      url: "http://localhost:8081/health"
      timeout: 30s
      interval: 1s
```

Сервисы стартуют и рестартуют строго в порядке списка. Останавливаются в обратном порядке.

---

## Промпт-стратегия

**Итерация 1:**
```
<user_prompt>

--- Test output ---
<test_output, truncated to max_test_output_lines>
```

**Итерация 2+ (обычная):**
```
--- Test output ---
<test_output, truncated>

--- Your changes from previous iterations ---
<git diff vs baseline_snapshot, truncated to max_diff_lines>
```

**Итерация 2+ (после отката):**
```
--- Test output ---
<test_output, truncated>

--- Previous attempt broke service restart ---
<git diff of rolled-back changes, truncated to max_diff_lines>

Try a different approach that doesn't break service startup.
```

---

## Логирование

Каждое событие выводится с цветным префиксом:

```
[INIT]    Starting svc-a...
[LIVE]    svc-a: waiting (3/30s)
[LIVE]    svc-a: OK ✓
[TEST]    Running test_command...
[TEST]    FAIL (exit 1)
[AGENT]   Running opencode... (timeout: 10m)
[AGENT] > Creating handler...        ← вывод агента
[AGENT] > Done.
[RESTART] Restarting svc-a...
[LIVE]    svc-a: OK ✓
[STOP]    Stopping svc-b...
[STOP]    Stopping svc-a...
```

---

## Расширяемость агента

Харнесс вызывает: `<command> <args...> <prompt>` — любой бинарник по этому соглашению. Промпт передаётся последним позиционным аргументом (строка, прочитанная из `prompt_file` + сгенерированный контекст). Агент завершается когда закончил, изменения на диске.

---

## Артефакты по завершению

- **Успех** (exit 0): изменения в рабочей директории, git-операции на усмотрение разработчика
- **Неудача** (exit 1): изменения в рабочей директории, харнесс печатает номер последней итерации и вывод теста

---

## Open questions

1. **`opencode run` exit codes** — не задокументированы, проверить эмпирически до имплементации.
2. **Инфраструктурные зависимости** — базы данных, очереди и прочее предполагаются уже запущенными локально. Харнесс не управляет ими.
