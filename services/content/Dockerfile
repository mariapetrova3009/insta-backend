# ---- build stage ----
FROM golang:1.22-alpine AS build
WORKDIR /src

RUN apk add --no-cache git ca-certificates build-base

# Сначала модули (лучший кеш)
COPY go.mod go.sum ./
RUN go mod download

# Дальше весь проект (чтобы видел internal/ и proto)
COPY . .

# Сборка бинарника сервиса content (путь к main.go)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o /out/content ./services/content/cmd

# ---- runtime stage ----
FROM gcr.io/distroless/base-debian12:nonroot
WORKDIR /
COPY --from=build /out/content /content
USER nonroot:nonroot
EXPOSE 9082
ENTRYPOINT ["/content"]
