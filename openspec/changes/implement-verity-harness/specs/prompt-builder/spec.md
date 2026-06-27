## ADDED Requirements

### Requirement: Промпт первой итерации
На первой итерации харнесс SHALL строить промпт из содержимого `prompt_file` и вывода теста.

#### Scenario: Построение промпта итерации 1
- **WHEN** это первая итерация цикла
- **THEN** промпт имеет формат:
  ```
  <содержимое prompt_file>

  --- Test output ---
  <test_output>
  ```

### Requirement: Промпт последующих итераций
На итерациях 2+ харнесс SHALL строить промпт из вывода теста и git diff от baseline-снапшота.

#### Scenario: Построение промпта итерации 2+
- **WHEN** это итерация 2 или выше без предшествующего rollback
- **THEN** промпт имеет формат:
  ```
  --- Test output ---
  <test_output>

  --- Your changes from previous iterations ---
  <git diff vs baseline_snapshot>
  ```

### Requirement: Rollback-промпт
Если предыдущая итерация завершилась rollback из-за liveness failure, харнесс SHALL строить специальный промпт с информацией об откате.

#### Scenario: Построение rollback-промпта
- **WHEN** предыдущая итерация завершилась rollback
- **THEN** промпт имеет формат:
  ```
  --- Test output ---
  <test_output>

  --- Previous attempt broke service restart ---
  <git diff of rolled-back changes>

  Try a different approach that doesn't break service startup.
  ```

### Requirement: Усечение diff в промпте
Харнесс SHALL обрезать git diff в промпте до `context.max_diff_lines` строк.

#### Scenario: Diff превышает лимит
- **WHEN** git diff содержит больше строк чем `max_diff_lines`
- **THEN** diff обрезается до `max_diff_lines` строк с пометкой об усечении
