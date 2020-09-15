#!/bin/bash

cd ../game
docker build -t qbed_game:test .

docker run --rm -itd --name game_client -v $PWD/data:/app/data qbed_game:test
docker run --rm -itd --name game_server -v $PWD/data:/app/data qbed_game:test

ovs-docker add-port server-br eth1 game_server --ipaddress=10.0.0.1/16
ovs-docker add-port client-br eth1 game_client --ipaddress=10.0.0.2/16
cd -
