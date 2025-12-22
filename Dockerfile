ARG GO_VERSION=1.24.0
FROM golang:${GO_VERSION}-alpine AS build
WORKDIR /data
USER root
ENV GOPROXY=https://goproxy.io,direct \
    CGO_ENABLED=0 \
    GOOS=linux
    
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -trimpath -ldflags "-s -w" -o /data/emnotonalBeach main.go

FROM alpine:latest AS final
WORKDIR /app
RUN apk add --no-cache ca-certificates tzdata && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone
ENV TZ=Asia/Shanghai
COPY --from=build /data/emnotonalBeach /app/
VOLUME /app/config
COPY entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh
EXPOSE 8080
ENTRYPOINT [ "./entrypoint.sh"]
