#!/usr/bin/env bash

go build

./control_kit  >/dev/null 2>&1 &

