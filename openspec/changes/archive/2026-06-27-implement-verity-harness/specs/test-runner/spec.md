## ADDED Requirements

### Requirement: Запуск тестовой команды
Харнесс SHALL выполнять `test_command` как shell-команду и захватывать объединённый stdout+stderr.

#### Scenario: Тест прошёл
- **WHEN** `test_command` завершается с exit code 0
- **THEN** харнесс считает тест зелёным и инициирует успешное завершение

#### Scenario: Тест упал
- **WHEN** `test_command` завершается с ненулевым exit code
- **THEN** харнесс захватывает вывод и передаёт в prompt-builder для следующей итерации

### Requirement: Усечение вывода теста
Харнесс SHALL обрезать захваченный вывод теста до `context.max_test_output_lines` строк (с конца — последние строки наиболее информативны).

#### Scenario: Вывод превышает лимит
- **WHEN** вывод теста содержит больше строк чем `max_test_output_lines`
- **THEN** в промпт передаются последние `max_test_output_lines` строк с пометкой об усечении

#### Scenario: Вывод в пределах лимита
- **WHEN** вывод теста меньше или равен `max_test_output_lines`
- **THEN** вывод передаётся целиком без изменений
