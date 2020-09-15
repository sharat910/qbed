#!/bin/bash

echo "Creating bridges..."
ovs-vsctl add-br server-br
ovs-vsctl add-br client-br
ovs-vsctl add-br switch
ovs-vsctl set bridge switch other_config:datapath-id=0000000000000001

echo "Creating ports and links..."
ip link add dev p1 type veth peer name swp1
ip link add dev p2 type veth peer name swp2

echo "Bringing up port interfaces..."
ip link set dev p1 up
ip link set dev p2 up
ip link set dev swp1 up
ip link set dev swp2 up

echo "Adding ports to bridges..."
ovs-vsctl add-port server-br p1
ovs-vsctl add-port client-br p2
ovs-vsctl add-port switch swp1
ovs-vsctl add-port switch swp2

echo "Done."

