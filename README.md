# RPA Collection

This project updates manual loan collectability and sets a loan's repayment saving account through Fincloud. Both CLI commands remain available, and the API exposes the same operations.

## Configuration

Copy `.env.example` to `.env` and fill in the Fincloud URLs and credentials. Programs load `.env` from the current directory; existing environment variables take precedence. `FINCLOUD_LOOKUP_LOCATION_ID` is needed for ten-digit alternate loan numbers, is required when starting the API, and must be allowed to search `cabang=ALL`. The Fincloud Web login for the mutation uses the branch derived from the resolved primary loan account. `FINCLOUD_LOCATION_ID` is no longer used. `RPA_API_KEY` is a separate credential for HTTP callers; it must not be the Fincloud API secret. The CLI retains the wrappers' default Fincloud URLs when the URL variables are blank; the API requires explicit URLs.

## CLI

```sh
go build -o app ./cmd
./app 1 Manual 3000010000000010 0130101394 3000020000000123

go build -o repayment-account ./cmd/autodebit
./repayment-account 3000010000000011 001000OPER
./repayment-account 0130101415 001123456789
```

The kolek command accepts `Manual` or `Automatic` and updates BI and BPR together. It skips only when both current values and change types already match. It resolves alternate accounts first, groups accounts by loan branch, and continues after an account or branch failure. Exit codes are `0` without failures, `1` with account failures, and `2` for invalid arguments or configuration. The repayment command accepts any active saving account returned by Fincloud saving balance inquiry, including cross-branch accounts and OPER accounts. It uses the requested saving account rather than the loan's disbursement account.

## HTTP API

```sh
go run ./cmd/api
```

The server listens on `RPA_LISTEN_ADDR` (default `:8080`). `GET /health` needs no token and does not contact Fincloud. Both mutation routes require `Authorization: Bearer $RPA_API_KEY` and `Content-Type: application/json`.

```sh
curl -X POST http://localhost:8080/api/v1/kolek \
  -H "Authorization: Bearer $RPA_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"kolek":5,"change_type":"Manual","accounts":["3000010000000010","0130101394"]}'

curl -X POST http://localhost:8080/api/v1/repayment-account \
  -H "Authorization: Bearer $RPA_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"loan_account":"3000010000000011","saving_account":"001000OPER"}'

curl -X POST http://localhost:8080/api/v1/repayment-account \
  -H "Authorization: Bearer $RPA_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"loan_account":"0130101415","saving_account":"001123456789"}'
```

Success responses use `{"status":"ok","data":...}`. Error responses use `{"status":"error","error":{"message":"..."}}`. Kolek returns HTTP 200 once a valid batch runs, including partial success; `data` contains `processed`, `success`, `skipped`, `failed`, and per-account `results`. Invalid JSON or business inputs return 400, invalid credentials 401, unsupported content type 415, and unsupported methods 405. Repayment returns 422 for a missing loan or saving account, a mismatched or incomplete saving inquiry, or an inactive saving account; Fincloud operation failures return 502 and server configuration failures return 500. Upstream response internals are not returned. JSON bodies are limited to 1 MiB.

For saving account inquiries, a non-success Fincloud API response is treated as “not found” only when its description says `not found` or `tidak ditemukan`; other upstream errors return 502. A loan inquiry with no result is treated as not found. These mappings follow the currently observed response shapes and may need adjustment if Fincloud changes them.
