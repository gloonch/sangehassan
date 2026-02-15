FROM golang:1.22-alpine AS base
WORKDIR /app
RUN apk add --no-cache git
COPY go.mod go.sum ./
# Some networks block proxy.golang.org. Use multiple proxies as fallback.
RUN GOPROXY='https://proxy.golang.org|https://goproxy.cn|direct' go mod download

FROM base AS dev
RUN go install github.com/cosmtrek/air@v1.50.0
COPY . .
CMD ["air", "-c", ".air.toml"]

FROM base AS build
COPY . .
RUN go build -o /bin/server ./cmd/server

FROM alpine:3.19 AS prod
WORKDIR /app
COPY --from=build /bin/server /app/server
COPY entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh
EXPOSE 8080
CMD ["/app/entrypoint.sh"]
