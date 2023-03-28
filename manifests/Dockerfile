FROM golang:1.19

LABEL "author"="fy403"
LABEL "version"="1.0"
LABEL "email"="1766056726@qq.com"

WORKDIR /usr/src/aurora
EXPOSE 4332
EXPOSE 8080

COPY . .
RUN go mod tidy --compat=1.17
RUN go install .
RUN rm -rf /usr/src/aurora/internal/...
# 默认启动web可替换
CMD ["aurora", "web", "--config", "/usr/src/aurora/config/config.toml"]