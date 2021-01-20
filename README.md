# yomo-sink-socketio

The example of [socket.io](https://socket.io/) for yomo-sink which can be used to show the real-time data on a web page.

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

1. **socket.io server**: accept the connections from socket.io clients (web pages) and broadcast the real-time data to clients.
2. **sink server**: receive the real-time data from `yomo-flow` and use [socket.io](https://socket.io/) to push the real-time data to web pages.

> BTW, you are [free to change the ports](https://github.com/yomorun/yomo-sink-socketio/blob/main/main.go#L15) of these two servers.


### Run `yomo-zipper`

You can use the command `yomo wf dev workflow.yaml` to run [yomo-zipper](https://yomo.run/zipper) which will automatically receive the real noise data from CELLA office, or run `yomo wf run workflow.yaml` with the specific `yomo-source`. See [yomo-zipper](https://yomo.run/zipper#how-to-config-and-run-yomo-zipper) for details.

After running `yomo-zipper`, you can visit `http://localhost:8000/public` in browser, it will show the data in real-time.

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
