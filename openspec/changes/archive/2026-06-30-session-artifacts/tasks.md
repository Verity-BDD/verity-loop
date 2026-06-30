## 1. Новый пакет `internal/session`

- [x] 1.1 Создать `internal/session/session.go` с типом `Session` (создаёт папку `.verity-sessions/<timestamp>/`, возвращает no-op при ошибке) и методом `Finish(outcome string, iterations int, duration time.Duration)` (пишет `session.md` и `session.json`)
- [x] 1.2 Добавить к `Session` метод `StartIteration(n int) *Iteration` — создаёт `iteration-<NN>/` и `iteration-<NN>/diffs/`, возвращает `*Iteration`
- [x] 1.3 Реализовать тип `Iteration` с методами: `WritePrompt(text string)`, `AgentWriter() io.WriteCloser`, `WriteTestOutput(text string)`, `WriteDiffs(diffs []snapshot.ServiceDiff)`, `WriteRollbackDiff(diffs []snapshot.ServiceDiff)`, `Finish(status string, duration time.Duration)`
- [x] 1.4 Написать unit-тест: успешная сессия создаёт ожидаемое дерево файлов; ошибка создания директории не паникует и возвращает no-op объекты

## 2. Tee вывода агента

- [x] 2.1 Изменить сигнатуру `agent.Runner.Run(ctx, prompt string, w io.Writer)` — добавить опциональный writer; в `readOutput` писать каждую строку и в logger, и в `w` (если `w != nil`)
- [x] 2.2 Обновить все call sites `agentRunner.Run(...)` в `harness.go` (пока передавать nil — подключим в следующем шаге)

## 3. Интеграция в `internal/harness`

- [x] 3.1 В начале `harness.Run()` создать `session.New(cfg.ConfigDir)` и `defer sess.Finish(...)`
- [x] 3.2 Внутри цикла итераций: вызывать `sess.StartIteration(i)` в начале каждой итерации
- [x] 3.3 Записывать дифф сервисов: после `baseline.DiffAll()` вызывать `iter.WriteDiffs(serviceDiffs)`
- [x] 3.4 Записывать промт: после `prompt.Build(...)` вызывать `iter.WritePrompt(promptStr)`
- [x] 3.5 Передавать `iter.AgentWriter()` в `agentRunner.Run(...)` для tee агентского вывода
- [x] 3.6 Записывать rollback-дифф: перед `preSnap.RestoreAll()` вызывать `iter.WriteRollbackDiff(rollbackDiffs)`
- [x] 3.7 Записывать тест-вывод: после `testrunner.Run(...)` вызывать `iter.WriteTestOutput(result.Output)`
- [x] 3.8 Записывать результат итерации: вызывать `iter.Finish(status, duration)` в конце каждой ветки (PASS, FAIL, ROLLBACK, TIMEOUT)

## 4. Финализация и тест корректности

- [x] 4.1 Убедиться что `sess.Finish()` вызывается во всех путях выхода из `harness.Run()` (early PASS, exhausted iterations, context cancel)
- [x] 4.2 Проверить что при SIGINT `session.md` и `session.json` записываются (graceful shutdown уже есть через context cancel)
- [x] 4.3 Прогнать `go test ./...` — убедиться что существующие тесты не сломаны изменением сигнатуры `agent.Run`
- [x] 4.4 Провести ручную проверку: запустить `verity-loop run` на примере `examples/hello-world`, проверить структуру `.verity-sessions/`
