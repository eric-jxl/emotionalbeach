ARG GO_VERSION=1.24.0
FROM golang:${GO_VERSION}-alpine AS build
WORKDIR /data
USER root
ENV GOPROXY=https://goproxy.io,direct \
    GOOS=linux \
    CGO_ENABLED=0

COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -trimpath -ldflags "-s -w" -o /data/emnotonalBeach main.go

FROM alpine:3.23.2 AS final
WORKDIR /app
ENV TZ=Asia/Shanghai
COPY --from=build /data/emnotonalBeach /app/

COPY entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh
EXPOSE 8080
ENTRYPOINT [ "./entrypoint.sh"]
