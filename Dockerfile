ARG GO_VERSION=1.24.0
FROM golang:${GO_VERSION}-alpine AS build
WORKDIR /data
USER root
ENV ENV GOPROXY=https://proxy.golang.org,direct\
    GOOS=linux \
    CGO_ENABLED=0

COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -trimpath -ldflags "-s -w" -o /data/emnotonalBeach main.go

FROM alpine:3.23.2 AS final
WORKDIR /app
ENV TZ=Asia/Shanghai
# tzdata 提供时区数据库，否则 Go/pgx 解析 Asia/Shanghai 会报 "unknown time zone"
RUN apk add --no-cache tzdata ca-certificates
COPY --from=build /data/emnotonalBeach /app/

COPY entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh
EXPOSE 8080
ENTRYPOINT [ "./entrypoint.sh"]
