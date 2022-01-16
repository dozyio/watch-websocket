# File Watcher & Websocket

A file watcher and websocket server written in Go.

Basic premise is that you want to watch for file system changes and inform a
Javascript client of the update. The JS client could then perform a reload of
the page - i.e. a laymans hot reload.

There are probably better solutions but I needed something lightweight without
using webpack.

The file watcher runs on the current working directory and the websocket on
**ws://127.0.0.1:12345/ws**

## Install
    git clone https://github.com/dozyio/watch-websocket
    cd watch-websocket
    go install

## Running

Run watch-websocket from the path you want to watch changes in.

    ./watch-websocket

## JS Client

Sample javascript client

    <script>
    var socket = new WebSocket("ws://127.0.0.1:12345/ws");
    var heartbeat;

    function wsHeartbeat() {
        socket.send('!');
    }

    function reload() {
        window.location.reload(true);
    }

    socket.onopen = function(event) {
        console.log('WS Connected');
        wsHeartbeat();
        heartbeat = setInterval(wsHeartbeat, 5000);
    };

    socket.onmessage = function(event) {
        console.log(`WS message ${event.data}`);
        if (event.data === 'reload') {
            // Add a small delay to allow files to save and other processes
            // to run (i.e. tailwind JIT)
            setTimeout(reload, 200);
        }
    };

    socket.onclose = function(event) {
        console.log('WS connection closed', event);
        clearInterval(heartbeat);
    };

    socket.onerror = function(error) {
        console.log(`WS error ${error.message}`);
        clearInterval(heartbeat);
    };
    </script>
