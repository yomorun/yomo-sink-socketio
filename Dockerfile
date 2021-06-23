FROM golang:buster

RUN apt-get update && \
    apt-get install nano iputils-ping telnet net-tools ifstat -y

RUN cp  /usr/share/zoneinfo/Asia/Shanghai /etc/localtime  && \
    echo 'Asia/Shanghai'  > /etc/timezone

WORKDIR $GOPATH/src/github.com/yomorun/yomo-sink-socketio-server-example
COPY . .
RUN go get -d -v ./...

EXPOSE 8000

CMD ["sh", "-c", "go run main.go"]
