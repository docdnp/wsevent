#!/bin/bash
PORT=${1:-8080}

docker run -it --network host --rm --name wsclient-$RANDOM alpine/curl:latest \
    --no-buffer      \
    --header "Connection: Upgrade"      \
    --header "Upgrade: websocket"      \
    --header "Host: example.com:80"      \
    --header "Origin: http://example.com:80"      \
    --header "Sec-WebSocket-Key: SGVsbG8sIHdvcmxkIQ=="      \
    --header "Sec-WebSocket-Version: 13" \
    http://127.0.0.1:$PORT/echo
