#!/bin/bash

echo "Deleting ports..."
ip link del port1
ip link del port2

echo "Deleting bridges..."
ovs-vsctl del-br internet
ovs-vsctl del-br nbn
ovs-vsctl del-br bng

echo "Done."
