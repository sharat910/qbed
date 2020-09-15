#!/bin/bash

docker kill `docker ps -q --filter "name=server"`
docker kill `docker ps -q --filter "name=client"`
