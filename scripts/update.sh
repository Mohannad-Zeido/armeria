#!/usr/bin/env bash

cd /opt/armeria

git pull

# client
if [ $1 = "client" ] || [ $1 = "both" ]; then
	cd /opt/armeria/client
	/usr/bin/yarn build
fi

# server
if [ $1 = "server" ] || [[ $1 = "both" ]]; then
	cd /opt/armeria
	/usr/local/go/bin/go build -o /opt/armeria/build/armeria ./cmd/armeria/main.go
fi