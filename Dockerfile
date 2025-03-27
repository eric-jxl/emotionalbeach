ARG GO_VERSION=1.23.7
FROM golang:${GO_VERSION}-alpine AS build
WORKDIR /data
USER root
ENV TZ=Asia/Shanghai
ENV GOPROXY=https://goproxy.io,direct
COPY . .
RUN go build -ldflags "-s -w" -o /data/emnotonalBeach main.go

FROM alpine:latest AS final
WORKDIR /app
COPY --from=build /data/emnotonalBeach /app/
VOLUME /app/config
COPY entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh
EXPOSE 8080
ENTRYPOINT [ "./entrypoint.sh"]
