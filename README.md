# onefetch

Onefetch is a simple, self-hostable Docker-based application for downloading files from 1fichier.com.
The sole purpose of this application is to be able to download files directly from the server where the application is hosted. (Example: downloading movies and series for Jellyfin)

> **Imortant**.  
> This application requires a 1fichier premium account API key.

## Content

- [frontend](./frontend//README.md)
- [backend](./backend//README.md)

# Deploy with docker

```yml
services:
  onefetch:
    build:
      context: .
      dockerfile: Dockerfile
    volumes:
      - ./data:/app/data # App data folder (db and logs)
      - ./downloads:/app/downloads # Downloaded files
    ports:
      - "3000:3000"
```
