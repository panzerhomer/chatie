#!/bin/bash

ITERATIONS=10

run_curl() {
    index=$1
    curl -X GET "http://localhost:3000/ws?name=user${index}" \
    -H 'Sec-WebSocket-Version: 13' \
    -H 'Sec-WebSocket-Key: xeAWiX32m1jNsWtAwaBTkQ==' \
    -H 'Connection: Upgrade' \
    -H 'Upgrade: websocket' \
    -H 'Sec-WebSocket-Extensions: permessage-deflate; client_max_window_bits' \
    -H 'Host: localhost:3000'
}

for ((i=0; i<ITERATIONS; i++)); do
    run_curl $i &  
done

wait