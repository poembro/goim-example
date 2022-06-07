FROM golang:1.17-alpine AS builder

COPY . /app
WORKDIR /app

ENV GO111MODULE=on
ENV GOPROXY="https://goproxy.cn,direct"

RUN echo "第一部分中，需要一个完整的go环境来编译我们的软件。注意第一部分的名称和别名builder" 

RUN rm -rf target/ 
RUN mkdir target/

COPY cmd/comet/comet-example.toml target/comet.toml
COPY cmd/logic/logic-example.toml target/logic.toml
COPY cmd/job/job-example.toml target/job.toml

RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-w -s" -v -o target/comet cmd/comet/main.go 
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-w -s" -v -o target/logic cmd/logic/main.go
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-w -s" -v -o target/job cmd/job/main.go

#####分界线####

FROM golang:alpine
LABEL maintainer="poembro@126.com" 
LABEL name="poembro/goim" version="0.0.1" description="这是一个golang goim kefu服务"


ENV PATH .:$PATH
ENV APP_NAME goim

WORKDIR /webser/go_wepapp/goim-demo   
RUN echo "第二部分中，只需要编译后的可执行文件。 所以跟换为更小的 alpine " 
COPY --from=builder /app/target /webser/go_wepapp/goim-demo/

EXPOSE 8000

#cmd命令 容器启动后默认执行的命令及其参数 但会被docker run 命令后面的命令行参数替换
CMD ["sh"]   

# entrypoint 容器启动时的执行命令 不会被忽略 一定会被执行
#ENTRYPOINT ["echo", "$name"]