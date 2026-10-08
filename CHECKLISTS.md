
# Part 1 
| Done | Pts | Requirement |
|:----:|:---:|-------------|
| [X] | 4 | `POST /api/shorten` returns **201** with `code` and `short_url` |
| [X] | 3 | **Idempotency:** second `POST` with the **same URL** returns the **same** `code` (test required) |
| [X] | 4 | `GET /{code}` returns **302** with correct `Location` |
| [X] | 3 | Unknown code → **404**; bad/missing URL → **400** |
| [X] | 3 | URL validation and no server-side fetch of long URL |
| [X] | 2 | Codes 6–8 chars for new URLs; collision strategy for **new** codes only |
| [X] | 2 | `-base` flag used for `short_url` |
| [X] | 2 | `httptest`: shorten + redirect; table tests for bad URL and unknown code |
| [X] | 2 | Concurrent test (include concurrent duplicate shorten for same URL); **`go test -race ./...`** passes |

# Part 2
| Done | Pts | Requirement |
|:----:|:---:|-------------|
| [X] | 5 | Metadata route **200** / **404** with correct JSON |
| [X] | 4 | `ErrNotFound`, `ErrInvalidURL` from store/domain |
| [X] | 4 | `%w` + `errors.Is` in HTTP mapping |
| [X] | 5 | `Store` interface + fake used in tests |
| [X] | 4 | Tests for metadata route and error mapping |
| [X] | 3 | Test or note in README: idempotency still works via `Store` / HTTP after Part 2 changes |

# Part 3
| Done | Pts | Requirement |
|:----:|:---:|-------------|
| [X] | 5 | Server timeouts configured |
| [X] | 5 | `Mutex` vs `RWMutex` matches behavior and `DECISIONS.md` |
| [X] | 5 | Benchmarks for shorten and redirect |
| [X] | 5 | README: benchmark line + profiling insight |
| [X] | 5 | Tests and `-race` still green |

# Part 4
| Done | Pts | Requirement |
|:----:|:---:|-------------|
| [X] | 6 | Persistent `Store` (GORM+DB or file-backed) |
| [X] | 5 | Startup load |
| [X] | 5 | Create persisted before response |
| [X] | 4 | Restart test (temp DB or temp files) |
| [X] | 3 | Config selects memory vs persistent store |
| [X] | 2 | `-race` clean with persistent store |

# Part 5
