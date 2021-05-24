# yomo-sink-socketio

The example of [socket.io](https://socket.io/) for [yomo-sink](https://yomo.run/sink) which can be used to show the real-time data on a web page.

## How to run the example

### 1. Run `yomo-zipper`

In order to experience the real-time data processing in YoMo, you can use the command `yomo wf dev workflow.yaml` to run [yomo-zipper](https://yomo.run/zipper) which will automatically receive the real noise data from YoMo office, or run `yomo wf run workflow.yaml` with the specific [yomo-source](https://yomo.run/source). See [yomo-zipper](https://yomo.run/zipper#how-to-config-and-run-yomo-zipper) for details.

### 2. Run `yomo-sink-socketio`

``` shell
go run main.go
```

You will see the following message:

```shell
2020/12/30 19:40:40 Starting socket.io server...
2020/12/30 19:40:40 ✅ Serving socket.io on 0.0.0.0:8000
2020/12/30 19:40:40 Connecting to zipper localhost:9000 ...
2020/12/30 19:40:40 ✅ Connected to zipper localhost:9000
```

It contains two steps:

1. Serve **socket.io server**: accept the connections from socket.io clients (web pages) and broadcast the real-time data to clients.
2. Connect to **yomo-zipper**: connect to [yomo-zipper](https://yomo.run/zipper), receive the real-time data from `yomo-zipper`, and broadcast it to socket.io clients.

> BTW, you are [free to change the ports](https://github.com/yomorun/yomo-sink-socketio/blob/main/main.go#L30) of the servers.

### 3. Receive real-time data on webpage

Visit `http://localhost:8000/public` in browser, it will show the data in real-time.

## How to receive and show the data on web page

You can find the example on [/asset/index.html](https://github.com/yomorun/yomo-sink-socketio/blob/main/asset/index.html).

- the `<script>` import

```html
<script src="/socket.io/socket.io.js"></script>
```

- NPM

```js
// ES6 import
import { io } from 'socket.io-client';
// CommonJS
const io = require('socket.io-client');
```

> **Note:** it doesn't support socket.io-client v3.

- Code snippet

```js
const socket = io();
// receive_sink is the event name of broadcast.
socket.on('receive_sink', function (msg) {
  $('#messages').append($('<li>').text(msg));
});
```

## How `yomo-sink-socketio` works

![YoMo](https://github.com/yomorun/yomo-sink-socketio/blob/main/yomo-sink.png)
