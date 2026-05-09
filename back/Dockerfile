ARG GO_IMAGE=golang:1.22-alpine
ARG ALPINE_IMAGE=alpine:3.19

FROM ${GO_IMAGE} AS base
WORKDIR /app
RUN apk add --no-cache git
COPY go.mod go.sum ./
ARG GOPROXY_URL=https://proxy.golang.org|https://goproxy.cn|direct
# Some networks block proxy.golang.org. Use multiple proxies as fallback.
RUN GOPROXY="${GOPROXY_URL}" go mod download

FROM base AS dev
RUN go install github.com/cosmtrek/air@v1.50.0
COPY . .
CMD ["air", "-c", ".air.toml"]

FROM base AS build
COPY . .
RUN go build -o /bin/server ./cmd/server

FROM ${ALPINE_IMAGE} AS prod
WORKDIR /app
COPY --from=build /bin/server /app/server
COPY entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh
EXPOSE 8080
CMD ["/app/entrypoint.sh"]
