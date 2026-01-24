FROM golang:1.22-alpine AS base
WORKDIR /app
RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN go mod download

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
EXPOSE 8080
CMD ["/app/server"]
