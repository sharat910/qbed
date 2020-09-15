#!/bin/bash

echo "Deleting ports..."
ip link del p1
ip link del p2

echo "Deleting bridges..."
ovs-vsctl del-br server-br
ovs-vsctl del-br client-br
ovs-vsctl del-br switch

echo "Done."
