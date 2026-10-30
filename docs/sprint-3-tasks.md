# Sprint 3: параллельная работа

База всех веток — `sprint-3`. Общий контракт —
[sprint-3-api-contract.md](sprint-3-api-contract.md). Не переносить сюда
альтернативную логику ЛР1–ЛР2.

| Участник и ветка | Свои файлы и работа | Переиспользование и тесты | Согласование / готовность |
| --- | --- | --- | --- |
| Соня — `sprint-3/sonya-like-api` | `internal/handler/likes.go`, собственные `internal/handler/likes_test.go`. Реализовать `LikeHandler(Reactions)` и `PairHandler(Pairs)`: ID, пустое тело PUT, DTO, `errors.Is/As`, HTTP-коды. | `usecase.ReactService.React`, `interaction.NewPair`, `Reader.Load`, `State.HasLike/IsMutual`. Проверить первый и повторный лайк, обратный лайк, чтение пустой пары, неверный вход, обёрнутые ошибки и 502/504. | С Ариной — получение ID из context, атомарный вызов React и регистрация маршрутов; с Олей — типы ошибок клиента. Готово: контрактные ответы, отсутствие фиктивного 501, все тесты проходят. |
| Ваня — `sprint-3/vanya-discovery-api` | `internal/handler/discovery.go`, собственные `internal/handler/discovery_test.go`. Реализовать `DiscoveryHandler(Discovery)`, query validation, interest, пагинацию и DTO. | `usecase.PairDiscoveryService.Discover`, `matching.NewDiscoveryService`, порядок `MatchCandidates`. Проверить фильтр, две страницы, пустой `items:[]`, пределы, повторы/неизвестные параметры, ошибки поиска. | С Ариной — context ID и локальный `ListProfiles`; с Олей — источник профиля текущего пользователя. Готово: фильтр до пагинации и стабильные страницы, нет 501. |
| Оля — `sprint-3/olya-profile-api` | `internal/profileclient/client.go`, `internal/profileclient/client_test.go`, `internal/profilemock/server.go`, `internal/profilemock/server_test.go`, `cmd/profile-api/main.go`. Реализовать `GetProfile`, пять режимов mock, отдельный сервер. | `usecase.EligibilityProfileReader`, `domain.Profile`, `domain.ErrProfileNotFound`. Проверить context, повторное использование клиента, 200/404/500, slow, large, JSON и неверный ID, размеры. | С Соней — `errors.Is/As` и маппинг 502/504; с Ариной — base URL и бюджеты. Готово: все пять режимов, лимит тела, сервер запускается отдельно. |
| Арина — `sprint-3/arina-runtime-store` | `internal/config/sprint3.go`, `internal/repository/memory/pair_store.go`, `cmd/linkup/main.go`, `internal/handler/middleware.go`, свои `*_test.go`. Реализовать environment config, сборку одного mux и сервера, X-Demo-User, X-Request-ID, slog, shutdown, pair-store и атомарность команд. | `usecase.NewReactService`, `usecase.NewPairDiscoveryService`, `matching.NewDiscoveryService`, `memory.NewProfiles`, `memory.NewUserRepository`, `interaction.Store`. Проверить middleware, shutdown и одновременные встречные/повторные лайки с `-race`. | С Соней и Ваней — context ID и маршруты; с Олей — конфигурация клиента/сервера; со всеми — атомарность React. Готово: один работающий LinkUp сервис, корректное завершение, нет потерянных лайков. |

`internal/handler/sprint3_contract.go`, `docs/sprint-3-api-contract.md`,
`docs/sprint-3-tasks.md`, `go.mod`, `internal/interaction/*`,
`internal/usecase/react.go`, `.github/workflows/ci.yml`, `.gitlab-ci.yml`
— общие файлы. Участники не редактируют их независимо без согласования:
сначала предложить изменение контракта команде, затем внести одной веткой.
Новые тестовые файлы каждый участник создаёт в своей зоне; фиктивные тесты
в каркас не добавлены.

Общие вопросы до завершения спринта: проверять ли существование `targetID`
для `GET /pairs`; передавать ли список профилей через отдельный HTTP endpoint
вместо локального `ListProfiles`; какие окончательные таймауты и тестовые ID
пяти режимов Profile API; как обеспечить атомарность `React` при сохранении
публичного интерфейса ЛР2.
