# yomo-sink-socketio

The example of [socket.io](https://socket.io/) for yomo-sink which can show the realtime data on a web page.

## How to run the example

``` shell
go run main.go
```

You will see the following message:

```shell
2020/12/30 19:40:40 Starting socket.io server...
2020/12/30 19:40:40 ✅ Serving socket.io on 0.0.0.0:8000
2020/12/30 19:40:40 Starting sink server...
2020/12/30 19:40:40 ✅ Listening on 0.0.0.0:4141
```

It contains two servers:

1. **socket.io server**: accept the connections from socket.io clients (web pages) and broadcast the realtime data to clients.
2. **sink server**: receive the realtime data from `yomo-flow` and use [socket.io](https://socket.io/) to push the realtime data to web pages.

When you run the [yomo-zipper](https://github.com/yomorun/yomo) and [yomo-source-demo](https://github.com/yomorun/yomo-source-demo), visit `http://localhost:8000/` in browser, it will show the data in realtime.

## How to receive and show the data on web page

```js
const io = require('socket.io-client');

const socket = io();
socket.on('receive_sink', function (msg) {
  $('#messages').append($('<li>').text(msg));
});
```

## how `yomo-sink-socketio` works

![YoMo](https://github.com/yomorun/yomo-sink-socketio/blob/main/yomo-sink.png)
