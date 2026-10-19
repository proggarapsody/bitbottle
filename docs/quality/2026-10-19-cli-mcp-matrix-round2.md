# Stage 2 — расширение публичной матрицы

Дата: 2026-10-19. Основа: `9134a0a7f916cf95aa778c62e770854311038375`, после разрешённого merge [#686](https://github.com/proggarapsody/bitbottle/pull/686). Предыдущий срез: [30/1158 проверенных пар](2026-10-19-cli-mcp-matrix.md).

**Полный Stage 2 не завершён. Релиз NOT READY.** Этот срез расширяет проверяемые семейства и исправляет обнаруженные дефекты; неизвестные операции и непроверенные параметры не исключаются из строгого gate.

## Что изменилось

- Публичные CLI-сценарии проверяют важные флаги начатых PR/branch операций и сохранность состояния после отклонённых записей. Используются реальные Cloud/Server adapters и HTTP-фикстуры с независимыми ожиданиями.
- MCP SDK/stdio сценарии расширены на базовые PR/repo/branch операции, затем commits/comments/tasks/tags. Их synthetic domain backends проверяют MCP-контракт; они не доказывают HTTP-совместимость с провайдером или live-поведение.
- Alias/config/completion проверяются как локальные операции: изменения читаются новым публичным вызовом, ошибочные операции сохраняют состояние, обращения к backend/HTTP/keyring запрещены тестовой фабрикой. Генерируемые completion scripts проверяются по формату выбранной shell, включая `--shell`/`-s`.
- `scope: local` даёт неприменимость Server/Cloud только после полного локального success/error доказательства и проверки важных входов. Производные N/A считаются отдельно от provider pass; каждая локальная операция добавляет одну локальную строку к двум исходным remote-строкам. Одной причины N/A недостаточно.
- Runner требует базовые assertions для каждого supported outcome: success — output/input/effect; error — error/input/effect. Контракт может добавить pagination и другие проверки, но не убрать минимум. Нестроковые причины unsupported/N/A отклоняются.

## Дефекты и границы доказательства

- `branch create --json` возвращал текст вместо структурированного результата. Публичная регрессия воспроизвела ошибку; исправление использует общий formatter и сохраняет обычный текстовый режим.
- Ручной merge в Cloud отправлял CLI-значение вместо API-стратегии. Нормализация проверяется по [официальному merge-контракту Cloud](https://developer.atlassian.com/cloud/bitbucket/rest/api-group-pullrequests/#api-repositories-workspace-repo-slug-pullrequests-pull-request-id-merge-post).
- Server/DC merge использовал неверное поле стратегии; запрос auto-merge отправлялся в endpoint повторной попытки. [REST-контракт DC 10.0](https://developer.atlassian.com/server/bitbucket/rest/v1000/api-group-pull-requests/) и [Javadoc 10.2](https://docs.atlassian.com/bitbucket-server/javadoc/10.2.0/api/com/atlassian/bitbucket/pull/PullRequestMergeRequest.html) разделяют запрос merge с `autoMerge`/`strategyId` и повторную попытку `/auto-merge`. Wire types исправлены в OpenAPI и сгенерированы; используются строгие регрессии, а не live DC.
- Создание Server-only BLOCKER task на Cloud могло сообщать успех после создания обычного комментария. MCP task list в Server мог возвращать обычные комментарии вместо пустого списка; теперь возвращается `[]`. Проверяется отказ до записи через текущую capability-модель. Это ограничение текущего adapter, а не заявление об отсутствии [Cloud task API](https://developer.atlassian.com/cloud/bitbucket/rest/api-group-pullrequests/#api-repositories-workspace-repo-slug-pullrequests-pull-request-id-tasks-post).
- Документированный Cloud auto-merge контракт и точная wire-стратегия Server для `--rebase` пока не подтверждены. Синтетический endpoint не считается доказательством. Соответствующие важные флаги остаются пробелами; неопределённость QA не превращается в новый продуктовый текст ошибки.

Server multiline inline comments этим срезом не сертифицируются: для Server проверен отказ при некорректном диапазоне без записи, для Cloud — успешный диапазон. Cloud `include_reactions` проверен как типизированная ошибка текущей capability, а не успешная загрузка реакций. Cloud task-list fallback отдельно проверен для `open`/`resolved`: возвращаются обычные комментарии без фильтра, исходные комментарии не меняются.

## Результат прогона

`make qa-matrix QA_MATRIX_DIR=/tmp/bitbottle-stage2-round2-verified` прошёл с race/shuffle без кеша: полный `./...`, Go exit 0, 0 ошибок, 0 пропусков обязательных тестов. Live acceptance явно `not_run` (1 opt-in skip); 3 пакета без тестов перечислены отдельно.

| Метрика | Результат |
| --- | ---: |
| Каталог | 298 runnable CLI + 281 MCP = 579 операций |
| Passing публичные сценарии | 243 |
| Проверенные требования | 59 операций, 125 reviewed rows |
| Полные строки | **123/1165** против прежних 30/1158 |
| Supported-строки | 99, включая 7 локальных строк |
| Unsupported-строки | 10 |
| Производные remote N/A по локальному доказательству | 14 |
| Незавершённые строки | **1042**: 1040 без reviewed-классификации и 2 начатые merge-пары |

Знаменатель вырос на 7 из-за отдельных локальных строк; каталог операций не изменился. N/A не является успешным вызовом провайдера, а число сценариев не является процентом качества продукта. Незавершённые reviewed-строки: Cloud PR merge (`--auto`, `--auto-off`, `--rebase`) и Server PR merge (`--rebase`). Один публичный Server task-list regression используется как дополнительное свидетельство; полного reviewed DoD этой операции ещё нет.

Строгая проверка сохранённого результата проверила hashes/source/events и ожидаемо завершилась exit 1: `matrix incomplete: 1042 provider pairs`. Источник имеет fingerprint `eb4201dd08498557f7407ca38c6926f3f868e2052e227760f5da1fe5f1db3213`. Это проверка текущих исходников, а не release artifact.

`make lint` — 0 проблем; 31 Python-регрессия runner прошла. Standards/spec review исправлений и классификации завершены без оставшихся блокеров этого среза. Дефекты воспроизведены до исправления; намеренные повреждения merge strategy и task severity ожидаемо роняли публичные проверки и затем были отменены. Raw events, catalog, manifest и report находятся в указанной локальной директории; CI публикует тот же набор как `qa-matrix-<run_id>-<attempt>`. Проверки remote SHA доступны в PR ветки.

## Остаток

Продолжить классификацию неизвестных операций и публичные success/error сценарии полного каталога; отдельно закрыть оставшиеся важные параметры и неподтверждённые provider contracts. Строгий gate сохраняет все незавершённые строки.

На финальном SHA нужны обязательный CI, Cloud acceptance и native/npm/Docker проверки конкретных артефактов с digest и связью с публикацией. Текущий hermetic прогон не заменяет эти gates. Release Please вновь завершился `Bad credentials` на основе `9134a0a` ([run 39198506910](https://github.com/proggarapsody/bitbottle/actions/runs/39198506910)); эта работа не меняет его credentials. [Общий остаток до релиза](2026-10-19-release-readiness.md).

**DC не проверен live.** Эта оговорка обязательна в QA и release notes; она не оправдывает найденные регрессии DC. Лицензия и инфраструктура DC не настраивались.
