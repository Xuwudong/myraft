#!/usr/bin/env bash

go build

host="localhost:"
server_addr=$host$2
client_addr=$host$3
id=$1
./myraft -id $1 -server_port $2 -client_port $3 >/dev/null 2>&1 &
if [ $id -ne 1 ]
then
  sleep 1
  curl -X POST http://localhost:3333/member_add -H 'Content-Type: application/json' -d '[{"member_id":'$id',"server_addr":"'$server_addr'","client_addr":"'$client_addr'"}]'
fi
