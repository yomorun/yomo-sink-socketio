FROM golang:buster

RUN apt-get update && \
    apt-get install nano iputils-ping telnet net-tools ifstat -y

RUN cp  /usr/share/zoneinfo/Asia/Shanghai /etc/localtime  && \
    echo 'Asia/Shanghai'  > /etc/timezone

RUN GO111MODULE=off go get github.com/yomorun/yomo; exit 0
RUN cd $GOPATH/src/github.com/yomorun/yomo && make install

WORKDIR $GOPATH/src/github.com/yomorun/yomo-sink-socketio-server-example
COPY . .
RUN go get -d -v ./...

EXPOSE 4141/udp 8003

CMD ["sh", "-c", "go run main.go"]