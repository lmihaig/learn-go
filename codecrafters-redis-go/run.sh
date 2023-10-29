#!/bin/bash
./spawn_redis_server.sh &

cd ./redis-tester
make test_with_redis
pkill -f spawn_redis_server.sh