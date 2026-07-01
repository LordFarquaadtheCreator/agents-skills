# Agent Instructions — appscript package

## Purpose

Encapsulates all communication with the deployed Google Apps Script web app. Commands in `cmd/` call methods on `AppScript` — they never deal with URLs, redirects, or payload construction.

## Files

- `appscript.go` — `AppScript` struct with `Get`, `Create`, `Patch`, `Delete` methods
- `utils.go` — config loading (`loadScriptURL`), `sheetsConfig` struct, `repoRoot` helper
- `appscript_test.go` — e2e tests using `httptest` mock server

## AppScript API

| Method | HTTP | Action field | Args |
|--------|------|-------------|------|
| `Get(params url.Values)` | GET | — | query params |
| `Create(entry map[string]interface{})` | POST | `create` | entry fields |
| `Patch(matchBy, update map[string]interface{})` | POST | `patch` | matchBy + update |
| `Delete(matchBy map[string]interface{})` | POST | `delete` | matchBy |

## 302 redirect handling

Apps Script returns 302 on POST. Go's default client converts POST→GET on redirect, dropping the body. `postFollowRedirect` captures the `Location` header, then GETs it to retrieve the actual response.

## Testing

`NewAppScriptWithURL(u string)` accepts a URL for test injection. Tests spin up `httptest.Server` mimicking the real 302 flow.

```bash
go test ./appscript/ -v
```
