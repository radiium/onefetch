# onefetch

Onefetch is a simple, self-hostable Docker-based application for downloading files from 1fichier.com.
The sole purpose of this application is to be able to download files directly to the server where the application is hosted — for example, downloading movies and series for Jellyfin.

> **Important**
> This application requires a 1fichier premium account API key.

## Content

- [frontend](./frontend/README.md)
- [backend](./backend/README.md)

## Deploy with docker

```yml
services:
  onefetch:
    image: radiium/onefetch:latest
    environment:
      - APP_API_URL_1FICHIER=https://api.1fichier.com/v1
      - APP_API_URL_JELLYFIN=http://your-jellyfin-host:8096
    volumes:
      - ./data:/app/data            # App data folder (db and logs)
      - ./downloads:/app/downloads  # Downloaded files
    ports:
      - "3000:3000"
```
