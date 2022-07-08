#!/bin/bash

echo "Creating bridges..."
ovs-vsctl add-br internet
ovs-vsctl add-br nbn
ovs-vsctl add-br bng
ovs-vsctl set bridge bng other_config:datapath-id=0000000000000001

echo "Creating ports and links..."
ip link add dev port1 type veth peer name swport1
ip link add dev port2 type veth peer name swport2

echo "Bringing up port interfaces..."
ip link set dev port1 up
ip link set dev port2 up
ip link set dev swport1 up
ip link set dev swport2 up

echo "Adding ports to bridges..."
ovs-vsctl add-port nbn port1
ovs-vsctl add-port internet port2
ovs-vsctl add-port bng swport1
ovs-vsctl add-port bng swport2

echo "Config TC..."
echo "Adding base latency..."
tc qdisc add dev swport2 root netem delay 20ms
tc qdisc add dev port2 root netem delay 20ms

echo "Adding queues in downstream..."
# basic config
# tc qdisc add dev swport1 root netem rate 25mbit limit 200

# two queue min-rate config
# Root qdisc
tc qdisc add dev swport1 root handle 1: htb default 12

# Root class
tc class add dev swport1 parent 1: classid 1:1 htb rate 5mbit

# Child classes
tc class add dev swport1 parent 1:1 classid 1:11 htb rate 1mbit ceil 5mbit
tc class add dev swport1 parent 1:1 classid 1:12 htb rate 4mbit ceil 5mbit

# Verify
tc class show dev swport1

# Filters
tc filter add dev swport1 protocol ip parent 1:0 prio 1 u32 match ip src 10.0.0.0/24 flowid 1:11
tc filter add dev swport1 protocol ip parent 1:0 prio 1 u32 match ip src 10.0.1.0/24 flowid 1:12

echo "Done."

