#!/bin/bash

sudo ./stop-dockers.sh
sudo ./net-down.sh
sudo ./net-up.sh
sudo ./start-dockers.sh