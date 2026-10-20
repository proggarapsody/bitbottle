# Качество критичных тестов: аудит 2026-10-20

Профиль покрывает **59 операций: 19 CLI и 40 MCP**, выбранных для **125 строк провайдеров**. Эти числа описывают учёт сценариев, а не силу assertions. Статически разобраны все прежние 19 обязательных тестов; это не доказательство исчерпывающей проверки каждого поведения. Повсеместных тавтологических assertions не обнаружено. Подмена внешней зависимости сама по себе не делает тест плохим.

## Что доказывали существующие проверки

Источник состава — [`release-critical.json`](../../test/qamatrix/release-critical.json); прежние результаты сохранены в [истории профиля](2026-10-19-critical-release-scenarios.md).

| Семейство прежних обязательных тестов | Количество | Наблюдаемая граница |
| --- | ---: | --- |
| Auth token | 3 | Factory и наблюдаемый вывод токена, приоритет окружения, ошибка отсутствующего токена |
| `TestScript/errfmt_auth` | 1 | Реальный процесс CLI и локальный HTTP fixture с 401; пользовательское оформление ошибки |
| Публичные CLI contract tests | 5 | Полное дерево команд → реальные адаптеры → локальный HTTP; файлы коммитов, пагинация веток, запрет удаления, текущий репозиторий, отказ Cloud auto-merge |
| MCP stdio/core | 2 | Публичный SDK/stdio → синтетический `FakeClient`; протокол, обработчик, вход, вывод и ошибки |
| Cloud acceptance security | 8 | Реальные границы `Run`, `Cleanup`, HTTP-клиента и окружения; ownership, очистка после ошибки, redirect, фильтрация credentials |

Публичные CLI-семейства используют [`public_matrix_integration_test.go`](../../test/contract/public_matrix_integration_test.go), [`public_critical_release_test.go`](../../test/contract/public_critical_release_test.go) и соседние тесты. Локальные alias/config/completion-сценарии проверяют собственное локальное поведение. MCP core/review/matrix-семейства в [`pkg/cmd/mcp`](../../pkg/cmd/mcp/) проверяют полезные handler/protocol-контракты, но их синтетический backend не подтверждает совместимость настоящего адаптера с HTTP провайдера. Прежний профиль требовал stdio/core; остальные семейства также входят в полный запуск.

Тесты [`internal/acceptancecloud`](../../internal/acceptancecloud/) не сводятся к декларациям меток: они выполняют рабочие границы с управляемыми внешними зависимостями. Их локальный результат всё равно не равен live acceptance.

## Что усилено

В обязательный список добавлены шесть тестов: **19 → 25**. Операции, провайдеры, important inputs, deferred inputs и остальная политика сохранены. [`qa_matrix.py`](../../scripts/qa_matrix.py) изменён только в утверждённом SHA256 профиля: `06b6fae0a5c7631465efba57f1b46f359a54883fb6428532e24525309eb1ff28`. Проверка структуры против `git HEAD` сохранена в `/tmp/bitbottle-critical-quality-cli/profile-additive-check.log`.

- [`public_critical_merge_validation_test.go`](../../test/contract/public_critical_merge_validation_test.go): `TestPublicCriticalPRMergeRejectsInvalidInput` проверяет конфликтующие флаги и отсутствующий/нечисловой ID без HTTP и успешного вывода. `TestPublicCriticalPRMergeFailurePreservesSourceBranch` проверяет 403/409, отсутствие удаления после отказа merge, OPEN PR и ветку через публичное чтение. Для 409 проверяется `pr.merge.conflict`; для 403 — `ErrPermission` и HTTP status, без фиксации нынешнего отсутствия специального dotted code.
- [`public_critical_http_integration_test.go`](../../pkg/cmd/mcp/public_critical_http_integration_test.go): `TestMCPPublicCriticalHTTP_CloudCommentLifecycle`, `TestMCPPublicCriticalHTTP_ServerCommentConcurrency`, `TestMCPPublicCriticalHTTP_OrdinaryMerge` проходят публичный MCP SDK/stdio и реальные адаптеры до HTTP fixture. Проверяются текст комментария, версия редактирования и подтверждённое состояние merge.
- [`report_security_test.go`](../../internal/acceptancecloud/report_security_test.go): `TestRun_SanitizesCredentialsFromReturnedAndPersistedReport` проверяет возвращённый отчёт, сохранённый JSON и текст возвращаемой ошибки с искусственными секретами.

Принятые рабочие ориентиры записаны в [`tdd/go-testing.md`](../../.agents/skills/tdd/go-testing.md), доступном через [README навыков](../../.agents/skills/README.md): публичное поведение, независимо заданные ожидаемые значения/состояния, небольшие таблицы похожих случаев по рекомендациям Uber и целевые отрицательные контроли. Это практические рекомендации, не обязательная новая архитектура тестов.

## Отрицательные контроли

Изменения дефектов выполнялись временными Go overlays; обычный production-файл для них не редактировался. PASS/FAIL ниже относится только к указанному набору. Полный critical runner запускает `./...`, включая unit-тесты. Поэтому PASS выбранного публичного семейства не означает пропуск дефекта всем репозиторием.

| Внесённый дефект | Прежние проверки | Новая проверка |
| --- | --- | --- |
| Cloud comment: пустой `content.raw` | Публичные MCP PASS; adapter unit FAIL | Cloud lifecycle FAIL |
| Server comment edit: версия 0 | Публичные MCP PASS; adapter unit FAIL | Server concurrency FAIL |
| Cloud merge: OPEN принят за завершённый | Публичные MCP PASS; adapter unit FAIL | Ordinary merge FAIL |
| Удалён guard конфликтующих ручных стратегий | Contract package PASS; `pkg/cmd/pr` FAIL | Invalid input FAIL |
| После ошибки merge удаляется source branch | Contract package PASS; `pkg/cmd/pr` PASS | Fail-safe FAIL |
| Любая ошибка merge проглочена | Contract package FAIL: существующая защита | Fail-safe FAIL |
| Identity sanitization отключена | Прежний missing-credential test PASS | Report security FAIL |

Это выборочная проверка чувствительности конкретных assertions, не оценка mutation coverage репозитория. Для последней строки не утверждается PASS всего acceptance-пакета.

Воспроизведение и журналы:

- CLI: `/tmp/bitbottle-critical-quality-cli/README.md`; команды, overlays и `validation-{old,new}.log`, `error-{old,new}.log`, `generic-error-old-positive-control.log`, `{validation,error}-old-cli-unit.log`.
- MCP: `/tmp/bitbottle-critical-quality-mcp/results.json` содержит точные аргументы `go test -race -overlay …` и коды завершения; соседние `*-existing.log`, `*-existing-adapter.log`, `*-new.log` содержат причины отказов.
- Security: `/tmp/bitbottle-critical-quality-security/overlay.json`, `overlay-mutation.log`, `baseline-before.log`, `regression-before-fix.log`, `fixed-targeted.log` и `package-race.log`. Контроль запускается через `go test -race -overlay /tmp/bitbottle-critical-quality-security/overlay.json ./internal/acceptancecloud -run 'TestRun_(SanitizesCredentialsFromReturnedAndPersistedReport|RejectsMissingCredentialAndWritesSanitizedReport)' -count=1`.

## Обнаруженный дефект и статус проверки

Обнаружен настоящий дефект: `Run` скрывал искусственные секреты в Report и JSON, но возвращал исходную ошибку с ними. [`acceptance/acceptance_test.go:47`](../../acceptance/acceptance_test.go#L47) печатает эту ошибку в журнал. Новый regression-тест воспроизвёл утечку до исправления (`regression-before-fix.log`). В [`run.go`](../../internal/acceptancecloud/run.go) добавлена очистка текста возвращаемой ошибки с сохранением unwrap identity: `errors.As` по-прежнему находит исходный `*url.Error`. Проверка прошла после исправления (`fixed-targeted.log`), затем прошёл весь security-пакет с `-race` (`package-race.log`): для настоящего дефекта сохранено доказательство red → green.

Локально прошли полный contract-пакет с `-race`, 37 Python-тестов runner и validator навыка; журналы — `restored-all.log`, `runner-tests.log`, `skill-validator.log` в CLI-каталоге доказательств.

Координатор выполнил `make qa-matrix QA_MATRIX_DIR=/tmp/bitbottle-critical-quality-final`, затем `make qa-critical-check` с тем же каталогом: обе команды прошли, критичный профиль подтвердил **125/125 строк и 25/25 обязательных тестов**. Это полнота выбранного учёта, не оценка качества assertions. `acceptance/TestAcceptance` остался явно `not_run` по предусмотренному исключению; live не запускался.

Bundle `/tmp/bitbottle-critical-quality-final/` фиксирует dirty local worktree на базе `2b3f0d5`. После прогона изменена только группировка MCP imports для goimports, без изменения поведения. Поэтому bundle является локальным доказательством до этой правки форматирования; окончательным подтверждением станет PR CI на чистом финальном head. Повторный `make lint` прошёл: **0 issues**, журнал `/tmp/bitbottle-critical-quality-lint.log`.

Server production-код не изменялся. **DC не проверен live**. Cloud auto-merge отложен; обязательный отказ сохранён. В этой работе live-проверки и релиз не выполнялись.
