# onefetch/backend

Onefetch backend

## Install

```bash
go mod download
go mod tidy
```

## Dev

```bash
go run main.go
```

## Build

```bash
go build -o onefetch-app .
```

## Env vars

| env var           | default       | description                             |
| ----------------- | ------------- | --------------------------------------- |
| APP_ENV           | `development` | `development` or `production` mode.     |
| APP_PORT          | `3000`        | exposed server port                     |
| APP_DOWNLOAD_PATH | `./downloads` | Basolute or relative path for downloads |
| APP_DATA_PATH     | `./data`      | Basolute or relative path for data      |

> [!TIP]
> In `development` mode, the frontend must be launched separately.
> In `production` mode, the frontend is embedded and served by the backend as a static file.
