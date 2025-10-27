#
# Build frontend
#

FROM node:22-alpine AS frontend-builder

WORKDIR /app/frontend
ENV PNPM_HOME="/pnpm"
ENV PATH="$PNPM_HOME:$PATH"
ENV CI=true
RUN corepack enable
COPY frontend/ ./
RUN --mount=type=cache,id=pnpm,target=/pnpm/store pnpm install --frozen-lockfile
RUN pnpm run build
#
# Build backend
#

FROM golang:1.25-alpine AS backend-builder

RUN apk add --no-cache gcc musl-dev sqlite-dev
WORKDIR /app
COPY backend/go.* ./
RUN go mod download
COPY backend/ ./
RUN CGO_ENABLED=1 GOOS=linux go build \
    # -ldflags="-w -s" \
    -trimpath \
    -o onefetch-app .

#
# Image finale
#

FROM alpine:latest

RUN apk --no-cache add ca-certificates sqlite-libs

RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

WORKDIR /app
COPY --from=backend-builder /app/onefetch-app .
COPY --from=frontend-builder /app/frontend/build ./web

RUN mkdir -p /app/data /app/downloads && \
    chown -R appuser:appuser /app
VOLUME ["/app/data"]
VOLUME ["/app/downloads"]
ENV APP_ENV=production
EXPOSE 3000
CMD ["./onefetch-app"]
