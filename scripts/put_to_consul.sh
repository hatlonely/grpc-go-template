#!/usr/bin/env bash

function load() {
    payload=$1
    path=$2
    curl -H "Content-Type:application/json" --request PUT --data-binary @${payload} http://localhost:8500/v1/kv/${path}
}

function main() {
    load configs/server/server.json grpc-go-template/configs/server/server.json
    load configs/client/client.json grpc-go-template/configs/client/client.json
}

main