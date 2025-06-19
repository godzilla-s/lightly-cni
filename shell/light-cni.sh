#!/bin/bash

cmd_add() {
    if [ "${IS_DEFAULT_GATEWAY}" == "true" ]; then
        IS_GATEWAY="true"
    fi 
}

# 建立并设置网桥
setup_bridge() {
    iplink = `ip link show ${BRIDGE_NAME}`
    if [ $? -ne 0 ]; then 
        ip link add ${BRIDGE_NAME} mtu ${MTU} type bridge
    fi 

    if [ "${PROMISC_MODE}" == "true" ]; then 
        ip link set dev ${BRIDGE_NAME} promisc on
    else 
        ip link set dev ${BRIDGE_NAME} promisc off
    fi 
}

setup_netns() {
    mkdir -p /var/run/netns
    ls -sfT ${CNI_NETNS} /var/run/netns/${CNI_CONTAINERID}
}

cleanup_netns() {
    rm -rf /var/run/netns/${CNI_CONTAINERID}
}

setup_veth() {
    host_veth=${1}

    ip netns exec ${CNI_CONTAINERID}
}
