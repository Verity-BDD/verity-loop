## 1. Инициализация проекта

- [x] 1.1 Создать Go-модуль (`go mod init`) и структуру директорий
- [x] 1.2 Добавить зависимости: YAML-парсер (`gopkg.in/yaml.v3`), цветной вывод (`github.com/fatih/color`)
- [x] 1.3 Настроить основной entrypoint `cmd/verity-harness/main.go` с командой `run`

## 2. Конфигурация (`config`)

- [x] 2.1 Определить структуры Go для `verity.yaml` (Config, Agent, Service, Liveness, Context)
- [x] 2.2 Реализовать загрузку и валидацию конфига с дефолтными значениями
- [x] 2.3 Реализовать проверку существования `prompt_file` при старте
- [x] 2.4 Написать тесты валидации конфига (обязательные поля, дефолты, отсутствие файла)

## 3. Логгер (`logger`)

- [x] 3.1 Реализовать структурированный логгер с цветными префиксами `[INIT]`, `[LIVE]`, `[TEST]`, `[AGENT]`, `[RESTART]`, `[STOP]`
- [x] 3.2 Реализовать стриминг вывода агента построчно с префиксом `[AGENT] >`
- [x] 3.3 Реализовать throttled logging для liveness polling (не чаще 1 раза в 5 секунд)

## 4. Git snapshot (`git-snapshot`)

- [x] 4.1 Реализовать `TakeSnapshot()` — сохранение текущего состояния рабочего дерева через patch-файл в temp-директории
- [x] 4.2 Реализовать `Diff(snapshot)` — вычисление diff между snapshot и текущим состоянием
- [x] 4.3 Реализовать `Restore(snapshot)` — применение обратного патча для rollback
- [x] 4.4 Написать тесты: snapshot → изменение → restore возвращает исходное состояние

## 5. Lifecycle сервисов (`service-lifecycle`)

- [x] 5.1 Реализовать запуск start-команды в фоне с отслеживанием PID
- [x] 5.2 Реализовать HTTP liveness polling с `liveness.interval` и `liveness.timeout`
- [x] 5.3 Реализовать выполнение stop/restart команд синхронно
- [x] 5.4 Реализовать `Teardown()` — остановка всех сервисов в обратном порядке с timeout 10s
- [x] 5.5 Реализовать перехват SIGINT/SIGTERM с вызовом `Teardown()`

## 6. Запуск тестов (`test-runner`)

- [x] 6.1 Реализовать выполнение `test_command` через shell, захват stdout+stderr
- [x] 6.2 Реализовать усечение вывода до `max_test_output_lines` строк (с конца)

## 7. Сборка промпта (`prompt-builder`)

- [x] 7.1 Реализовать построение промпта итерации 1 (prompt_file + test output)
- [x] 7.2 Реализовать построение промпта итераций 2+ (test output + diff от baseline)
- [x] 7.3 Реализовать построение rollback-промпта (test output + "broke service" + diff отката)
- [x] 7.4 Реализовать усечение diff до `max_diff_lines` строк

## 8. Запуск агента (`agent-runner`)

- [x] 8.1 Реализовать запуск агента через `exec.Command(command, args..., prompt)` без shell
- [x] 8.2 Реализовать стриминг stdout агента в логгер построчно
- [x] 8.3 Реализовать таймаут агента с kill subprocess по истечении `agent.timeout`
- [x] 8.4 Реализовать счётчик `consecutive_timeouts` и выход при достижении 3

## 9. Главный цикл (`harness-loop`)

- [x] 9.1 Реализовать INIT-фазу: baseline snapshot → запуск сервисов → liveness → предварительный тест
- [x] 9.2 Реализовать основной LOOP: pre-agent snapshot → тест → промпт → агент → restart → liveness → проверка rollback
- [x] 9.3 Реализовать логику rollback при liveness failure после restart
- [x] 9.4 Реализовать завершение: exhausted (exit 1) и success (exit 0) с teardown

## 10. Интеграционное тестирование

- [x] 10.1 Написать E2E тест: харнесс находит зелёный тест с первой попытки → exit 0
- [x] 10.2 Написать E2E тест: mock-агент делает изменение → тест зеленеет на итерации 2
- [x] 10.3 Написать E2E тест: исчерпание `max_iterations` → exit 1
- [ ] 10.4 Проверить работу с реальным `opencode run` на простом примере
