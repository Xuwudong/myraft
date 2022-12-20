#!/usr/bin/env bash

go build

cd client
go build
cd ..
#
if ["$id" = ""]; then
  ./myraft >/dev/null 2>&1 &
  sleep 1
  cd client
  ./client_kit  2>&1 &
else
  ./myraft -id $id -server_port $2 -client_port $3 >/dev/null 2>&1 &
  sleep 1
  cd client
  ./client_kit -id $id -server_port $2 -client_port $3 2>&1 &
fi
