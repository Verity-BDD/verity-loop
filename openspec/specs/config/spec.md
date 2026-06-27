## ADDED Requirements

### Requirement: Загрузка verity.yaml
Харнесс SHALL загружать конфигурацию из файла `verity.yaml` в текущей рабочей директории при запуске.

#### Scenario: Успешная загрузка
- **WHEN** в рабочей директории существует валидный `verity.yaml`
- **THEN** харнесс загружает конфигурацию и продолжает работу

#### Scenario: Файл не найден
- **WHEN** `verity.yaml` отсутствует в рабочей директории
- **THEN** харнесс завершается с exit 1 и сообщением об ошибке

### Requirement: Обязательные поля конфигурации
Конфигурация SHALL содержать: `agent.command`, `test_command`, `prompt_file`, хотя бы один сервис в `services`.

#### Scenario: Отсутствует обязательное поле
- **WHEN** в `verity.yaml` отсутствует любое из обязательных полей
- **THEN** харнесс завершается с exit 1 и указанием какого поля не хватает

### Requirement: Дефолтные значения
Харнесс SHALL применять дефолты: `max_iterations: 10`, `agent.timeout: 10m`, `context.max_diff_lines: 200`, `context.max_test_output_lines: 100`, `liveness.interval: 1s`.

#### Scenario: Поле не указано в конфиге
- **WHEN** опциональное поле отсутствует в `verity.yaml`
- **THEN** харнесс использует дефолтное значение без ошибки

### Requirement: Валидация prompt_file
Харнесс SHALL проверять существование файла, указанного в `prompt_file`, до начала цикла.

#### Scenario: prompt_file не существует
- **WHEN** путь в `prompt_file` указывает на несуществующий файл
- **THEN** харнесс завершается с exit 1 до запуска сервисов

### Requirement: Переменные окружения сервиса
Сервис SHALL поддерживать опциональное поле `env` — map строк, передаваемых как переменные окружения в start/stop/restart команды.

#### Scenario: Env vars инжектируются в команду
- **WHEN** у сервиса задан `env: {KEY: value}`
- **THEN** переменная `KEY=value` присутствует в окружении при запуске start/stop/restart команд этого сервиса
