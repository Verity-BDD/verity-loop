## Why

Разработчику нужен инструмент, который замыкает петлю между падающим acceptance-тестом и LLM-агентом: запускает сервисы, запускает тест, кормит вывод агенту, итерирует до зелёного. Без такого харнесса этот цикл выполняется вручную и медленно.

## What Changes

- Новый CLI-бинарник `verity-harness` с командой `run`
- Конфигурация через `verity.yaml`: список сервисов, тестовая команда, агент, промпт-файл
- Lifecycle управления сервисами: start → liveness poll → restart после каждой итерации → stop при выходе
- Итеративный цикл: test → build prompt → agent → restart → repeat
- Промпт-стратегия: user prompt + test output на итерации 1, diff + test output на итерациях 2+, специальный rollback-промпт если сервис сломался после изменений агента
- Откат git-изменений агента если liveness упал после restart
- Логирование с цветными префиксами `[INIT]`, `[LIVE]`, `[TEST]`, `[AGENT]`, `[RESTART]`, `[STOP]`
- Таймаут агента (дефолт 10m), счётчик consecutive timeouts (3 подряд → exit 1)

## Capabilities

### New Capabilities

- `config`: Загрузка и валидация `verity.yaml` — агент, сервисы, тестовая команда, промпт-файл, контекстные лимиты
- `service-lifecycle`: Управление сервисами — start/stop/restart с liveness polling и teardown при выходе
- `test-runner`: Запуск `test_command`, захват stdout/stderr, определение pass/fail
- `prompt-builder`: Сборка промпта для каждой итерации из prompt_file + test output + git diff
- `agent-runner`: Запуск агента как subprocess, стриминг вывода, таймаут, счётчик consecutive timeouts
- `git-snapshot`: Снимок git-состояния в начале запуска и перед каждым агентом, откат при liveness failure
- `harness-loop`: Оркестрация полного цикла: INIT → LOOP → teardown, обработка SIGINT/SIGTERM
- `logger`: Цветной структурированный вывод с префиксами

### Modified Capabilities

## Impact

- Новый Go-модуль/бинарник, нет зависимостей от существующего кода
- Внешние зависимости: `opencode` (или любой совместимый агент), `go test`, `git`
- Требует git-репозитория в рабочей директории (для снапшотов и diff)
