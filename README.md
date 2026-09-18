# Update Manual Kolektibilitas

Build and run:

```sh
cp .env.example .env
# Fill in Fincloud URL and credentials in .env.
go build -o app ./cmd
./app 1 3000010000000010 0130101394 3000020000000123
```

The command loads `.env` from its current working directory with `godotenv`. Existing environment variables take precedence. `.env` is ignored by Git; `.env.example` lists required keys. Branch code from the primary account is used directly as the Fincloud login location ID. `FINCLOUD_LOOKUP_LOCATION_ID` is required only for ten-digit alternate accounts; that location must be allowed to search `cabang=ALL`. All accounts use `FINCLOUD_USERNAME`, `FINCLOUD_PASSWORD`, and `FINCLOUD_ROLE_ID`.

The command resolves alternate accounts first, then processes branches in code order. Accounts within a branch retain input order. One failed account does not stop the batch. A failed branch login fails that branch's accounts and processing continues. Exit status is `0` without failures, `1` with account failures, or `2` for invalid arguments or configuration. A skipped account is not a failure.
