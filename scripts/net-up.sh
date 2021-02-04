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
tc qdisc add dev port2 root netem delay 20ms
tc qdisc add dev swport1 root netem rate 25mbit limit 200

echo "Done."

