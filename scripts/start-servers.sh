#!/bin/bash

cd ../game || exit

docker build -t qbed_game:test .

docker run --rm -itd --name game_server \
          -v $PWD/data/server:/app/data \
          -v $PWD/logs/server:/app/logs \
          -p 8888:8888/udp \
          qbed_game:test

cd - || exit

cd ../download || exit
docker build -t qbed_download:test .

docker run --rm -itd --name download_server \
          -v $PWD/data/client:/app/data \
          -v $PWD/logs/client:/app/logs \
          -p 9999:9999 \
          qbed_download:test

cd - || exit

