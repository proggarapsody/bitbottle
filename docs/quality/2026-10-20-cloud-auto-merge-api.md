# Bitbucket Cloud: auto-merge, очереди и API-аутентификация

**Дата проверки:** 2026-10-20. Это заметка о текущих публичных API и полученных диагностических результатах; API-поведение, которого не проверили live, помечено отдельно.

## Публичные интерфейсы

Официальный REST API документирует `POST /2.0/repositories/{workspace}/{repo_slug}/pullrequests/{pull_request_id}/merge`. У него есть query `async` и поля тела `type`, `message`, `close_source_branch`, `merge_strategy`. Отдельного флага «выполнить merge после прохождения сборок» в этой операции не описано. Если для целевой ветки включена merge queue, этот POST помещает PR в очередь и возвращает `202 Accepted` с ссылкой для опроса результата; это поведение описано для активной очереди. [REST: Merge a pull request](https://developer.atlassian.com/cloud/bitbucket/rest/api-group-pullrequests/)

Публичный REST API также документирует merge queues: список очередей, получение очереди и PR в очереди, удаление PR из очереди через `DELETE /repositories/{workspace}/{repo_slug}/merge-queues/{merge_queue_uuid}/queue/{pullrequest_id}`. Удаление асинхронное и возвращает `202 Accepted`. [REST: Other operations](https://developer.atlassian.com/cloud/bitbucket/rest/api-group-other-operations/)

Merge queues — отдельная функция с требованиями: платный план Standard или Premium, включённые Bitbucket Pipelines, настроенная очередь и соответствующая секция `merge-queues` в `bitbucket-pipelines.yml`. При таком процессе Bitbucket строит и проверяет merge result перед обновлением целевой ветки. Поэтому очередь не является прямой заменой старому сценарию «ожидать статус уже существующей сборки» и требует отдельной инфраструктуры. [Atlassian Support: Manage pull requests with merge queues](https://support.atlassian.com/bitbucket-cloud/docs/manage-pull-requests-with-merge-queues/)

Проверенная официальная OpenAPI-спецификация [swagger.v3.json](https://dac-static.atlassian.com/cloud/bitbucket/swagger.v3.json?_v=2.300.200) содержит merge queues, но не содержит строк `post-build-merge` или `pending_merge`. Это свидетельство только об опубликованной спецификации, не доказательство отсутствия внутреннего или иного API.

## Аутентификация и следующий диагностический шаг

В объявлении Atlassian от 2026-02-09 указано, что с 2026-05-04 token-authenticated запросы должны направляться на `api.bitbucket.org`, а OAuth access token передаётся только в заголовке `Authorization: Bearer`. [Объявление Atlassian о смене OAuth/API-аутентификации](https://community.developer.atlassian.com/t/oauth-2-0-and-api-authentication-changes-for-bitbucket-cloud/99003)

Hosted run [39346130003](https://github.com/proggarapsody/bitbottle/actions/runs/39346130003) проверил actor- и bootstrap-учётные данные с Bearer и Basic на `https://bitbucket.org/!api/internal/.../post-build-merge`; все четыре запроса получили 401. Этот адрес использовал неверный для token-auth origin и внутренний UI-маршрут. Поэтому результат не доказывает несовместимость токенов с правильным API-origin. Старый путь `pending_merge` также оказался устаревшим.

Hosted run [39349570133](https://github.com/proggarapsody/bitbottle/actions/runs/39349570133) проверил правильный origin и endpoint. Публичные PR GET/PUT ответили 200; четыре внутренних GET (actor/bootstrap, Bearer/Basic) ответили 403, а actor POST ответил 400: `The destination branch doesn't support automatically merging when builds pass.` Outcome остался `unconfirmed`; cleanup прошёл, а `CLEANUP_STATE` был независимо проверен как `verified`. Эти результаты не подтверждают token-native работу маршрута: первый диагностический PR был конфликтным, а репозиторию не хватало включённой branch configuration.

Hosted run [39351214138](https://github.com/proggarapsody/bitbottle/actions/runs/39351214138) не дошёл до внутренних auto-merge запросов. Branch rules и создание ветки ответили 201, затем чтение commit через `GET /commit/auto-probe%2F{runID}` ответило 404; cleanup прошёл. Это infrastructure failure подготовки fixture, не результат проверки auto-merge API. Исправленный probe получает commit hash из `GET /refs/branches/{escaped_branch}` и `target.hash`; token-native flow остаётся непроверенным до следующего hosted run.

Следующий диагностический шаг создаёт отдельный непересекающийся PR в том же временном репозитории. Он добавляет официальные правила `allow_auto_merge_when_builds_pass` и `require_passing_builds_to_merge` только для `main`, ставит синтетический build status и проверяет POST, отмену через DELETE, затем повторный POST с успешным статусом. Правила создаются через официальный [Branch restrictions API](https://developer.atlassian.com/cloud/bitbucket/rest/api-group-branch-restrictions/). Этот flow ограничен disposable repository; он не меняет продукт и сам по себе не является acceptance или CLI pass.

Отдельно, полный browser-flow на [изолированном PR 1](https://bitbucket.org/bitbucket-qa-lab/acceptance-sandbox/pull-requests/1) доказал создание ожидания (201), отмену (204) и автоматический merge после синтетического `SUCCESSFUL`; PR был объединён в собственную целевую ветку, основная ветка не менялась. Это не проверило CLI-токены или фактический запуск сборки.

Локальная проверка Keychain также возвращала 401 для обоих старых вариантов. Действующие диагностические токены доступны только в GitHub Environment, ограниченном `main`; данных пока недостаточно, чтобы утверждать, что нужны их перевыпуск или другие scopes. `auto-merge` ранее отвечал 404; текущие CLI-флаги остаются неподтверждёнными. Продукт не менялся, выпуск заблокирован; **DC не проверен live**.

## Статус реализации, 2026-10-20

Продуктовый контракт для Cloud изменён на официальный merge queue: `pr merge --auto` требует активную очередь целевой ветки, отправляет выбранную merge strategy через публичный `POST .../pullrequests/{id}/merge` и проверяет появление PR в очереди либо его завершённый merge. Без активной очереди команда возвращает `pr.merge.queue_required` до отправки merge-запроса. `--auto-off` ищет PR в очередях, включая `DRAINING`, удаляет его и проверяет отсутствие при открытом PR. Это не трактует прежний `post-build-merge` диагностический результат как pass и не меняет вывод о прежнем `/auto-merge` маршруте.

Проверки HTTP-контрактов и hermetic acceptance не подтверждают live поведение Bitbucket. Официальная Cloud queue acceptance описана в `docs/workflows/acceptance.md`; сейчас live pass не заявляется: QA workspace пока не позволяет завершить настройку очереди, а workflow также требует разрешения `read:pipeline:bitbucket` для bootstrap credential в своём read-only preflight. Server/DC остаётся на отдельном классическом поле `autoMerge` при `POST .../pull-requests/{id}/merge`; его live-поведение этой Cloud проверкой не покрывается.
