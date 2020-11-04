#!/bin/bash

scp x3:~/qbed/expts/pcaps/"$1".pcap.gz ../expts/tags/"$1"/
scp -r x3:~/qbed/game/data/client/"$1" ../expts/tags/"$1"/