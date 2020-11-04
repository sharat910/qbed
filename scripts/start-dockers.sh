#!/bin/bash

cd ../game || exit

docker build -t qbed_game:test .

docker run --rm -itd --name game_server \
          -v $PWD/data/server:/app/data \
          -v $PWD/logs/server:/app/logs \
          --cap-add=NET_ADMIN \
          qbed_game:test
docker run --rm -itd --name game_client \
          -v $PWD/data/client:/app/data \
          -v $PWD/logs/client:/app/logs \
          --cap-add=NET_ADMIN \
          qbed_game:test


ovs-docker add-port internet eth1 game_server --ipaddress=10.0.0.1/16
ovs-docker add-port nbn eth1 game_client --ipaddress=10.0.0.2/16
cd - || exit

cd ../download || exit
docker build -t qbed_download:test .

docker run --rm -itd --name download_client \
          -v $PWD/data/client:/app/data \
          -v $PWD/logs/client:/app/logs \
          --cap-add=NET_ADMIN \
          qbed_download:test
docker run --rm -itd --name download_server \
          -v $PWD/data/server:/app/data \
          -v $PWD/logs/server:/app/logs \
          --cap-add=NET_ADMIN \
          qbed_download:test

ovs-docker add-port internet eth1 download_server --ipaddress=10.0.1.1/16
ovs-docker add-port nbn eth1 download_client --ipaddress=10.0.1.2/16
cd - || exit
