## ADDED Requirements

### Requirement: Вызов агента
Харнесс SHALL вызывать агента командой `<agent.command> <agent.args...> <prompt>`, где prompt — последний позиционный аргумент. Вывод агента стримится в лог с префиксом `[AGENT] >`.

#### Scenario: Агент завершился успешно
- **WHEN** subprocess агента завершается с exit code 0
- **THEN** харнесс считает итерацию выполненной, сбрасывает счётчик consecutive_timeouts, переходит к restart

#### Scenario: Агент завершился с ошибкой (ненулевой exit code)
- **WHEN** subprocess агента завершается с ненулевым exit code
- **THEN** харнесс логирует предупреждение, считает итерацию выполненной (изменения могли быть сделаны частично), переходит к restart

### Requirement: Таймаут агента
Харнесс SHALL завершать subprocess агента по истечении `agent.timeout` (дефолт: 10m).

#### Scenario: Агент завис
- **WHEN** агент не завершается в течение `agent.timeout`
- **THEN** харнесс убивает subprocess, инкрементирует `consecutive_timeouts`, пропускает restart и переходит к следующей итерации без изменений

### Requirement: Лимит последовательных таймаутов
Харнесс SHALL завершаться с exit 1 если агент завис `consecutive_timeouts >= 3` раз подряд.

#### Scenario: Три таймаута подряд
- **WHEN** агент не завершился в течение таймаута три итерации подряд
- **THEN** харнесс вызывает teardown и завершается с exit 1 с сообщением "agent timed out 3 times in a row"

#### Scenario: Сброс счётчика при успехе
- **WHEN** агент успешно завершается после одного или двух таймаутов
- **THEN** счётчик `consecutive_timeouts` сбрасывается в 0
