#!/bin/sh
nohup go run merchant/rpc/merchant-rpc.go 2>&1 &
nohup go run encryption/rpc/encryption-rpc.go 2>&1 &
nohup go run redis/rpc/redis-rpc.go 2>&1 &
nohup go run session/rpc/session-rpc.go 2>&1 &
nohup go run merchant/gateway/merchant-rpc.gw.go 2>&1 &
nohup go run session/gateway/session-rpc.gw.go 2>&1 &

