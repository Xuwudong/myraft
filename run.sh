#!/usr/bin/env bash

if ["$1" = ""]; then
  ./myraft > /dev/null 2>&1 &
else
  ./myraft -id $1 > /dev/null 2>&1 &
fi
