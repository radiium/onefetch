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

| env var              | default                        | description                             |
| -------------------- | ------------------------------ | --------------------------------------- |
| APP_ENV              | `development`                  | `development` or `production` mode      |
| APP_PORT             | `3000`                         | Exposed server port                     |
| APP_DOWNLOAD_PATH    | `./downloads`                  | Absolute or relative path for downloads |
| APP_DATA_PATH        | `./data`                       | Absolute or relative path for data      |
| APP_API_URL_1FICHIER | `https://api.1fichier.com/v1`  | 1fichier API base URL                   |
| APP_API_URL_JELLYFIN | `http://192.168.1.20:8096`     | Jellyfin API base URL                   |

> [!TIP]
> In `development` mode, the frontend must be launched separately.
> In `production` mode, the frontend is embedded and served by the backend as a static file.

## API

The full API is documented in [`specs/openapi.yml`](../specs/openapi.yml).
