FROM golang:1.19

LABEL "author"="fy403"
LABEL "version"="1.0"
LABEL "email"="1766056726@qq.com"

WORKDIR /usr/src/app
EXPOSE 4332
EXPOSE 8080
# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -v -o /usr/local/bin/aurora ./...
# 默认启动web可替换
CMD ["aurora", "web", "--config", "./config/config.toml"]