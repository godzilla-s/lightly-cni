# lightly-cni
lightly-cni is a lightweight CNI fork3s/kubernetes


## 背景知识

如何手动创建一个CNI所支持的网络： 

1. 创建网络命名空间

```
ip netns add test.0
```

创建完成后: 

```shell
root@localhost:~/k3s# ip netns ls
test.0
```

2. 创建veth pair

其作用主要是用于容器内部网络与主机的通信， 

其特点：

+ 成对出现：数据从一端进入，会从另一端发出（类似管道）。
+ 跨命名空间通信：常用于连接不同网络命名空间（如容器和主机）。
+ 无物理实体：纯软件实现的虚拟设备。

创建: 

```
ip link add veth0 type veth peer name veth1
```

查看：

```shell
root@localhost:~/k3s# ip addr
22: veth1@veth0: <BROADCAST,MULTICAST,M-DOWN> mtu 1500 qdisc noop state DOWN group default qlen 1000
    link/ether b2:f4:f2:2d:1d:63 brd ff:ff:ff:ff:ff:ff
23: veth0@veth1: <BROADCAST,MULTICAST,M-DOWN> mtu 1500 qdisc noop state DOWN group default qlen 1000
    link/ether a6:20:4e:00:df:4b brd ff:ff:ff:ff:ff:ff
```

3. 将其中一个虚拟设备加入到网络命名空间

```
ip link set veth1 netns test.0
```

查看:

```shell
root@localhost:~/k3s# ip netns exec test.0 ip addr
1: lo: <LOOPBACK> mtu 65536 qdisc noop state DOWN group default qlen 1000
    link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00
22: veth1@if23: <BROADCAST,MULTICAST> mtu 1500 qdisc noop state DOWN group default qlen 1000
    link/ether b2:f4:f2:2d:1d:63 brd ff:ff:ff:ff:ff:ff link-netnsid 0
```

启用:

```
ip netns exec test.0 ip link set veth1 up
```
查看：
```shell
root@localhost:~/k3s# ip netns exec test.0 ip addr
1: lo: <LOOPBACK> mtu 65536 qdisc noop state DOWN group default qlen 1000
    link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00
22: veth1@if23: <NO-CARRIER,BROADCAST,MULTICAST,UP> mtu 1500 qdisc noqueue state LOWERLAYERDOWN group default qlen 1000
    link/ether b2:f4:f2:2d:1d:63 brd ff:ff:ff:ff:ff:ff link-netnsid 0
```

通样将另一端也启用：
```
ip link set veth0 up
```

4. 配置IP地址

```
ip netns exec test.0 ip addr add 10.40.0.10/24 dev veth1
```
查看：
```shell
root@localhost:~/k3s# ip netns exec test.0 ip addr
1: lo: <LOOPBACK> mtu 65536 qdisc noop state DOWN group default qlen 1000
    link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00
22: veth1@if23: <NO-CARRIER,BROADCAST,MULTICAST,UP> mtu 1500 qdisc noqueue state LOWERLAYERDOWN group default qlen 1000
    link/ether b2:f4:f2:2d:1d:63 brd ff:ff:ff:ff:ff:ff link-netnsid 0
    inet 10.40.0.10/24 scope global veth1
       valid_lft forever preferred_lft forever
```

5. 配置路由

如果要实现不同容器的通信，需要配置路由。

为了测试可以按照上面方法创建容器网络空间`test.1`，以及对应的veth pair（`veth2`，`veth3`），并配置IP地址 10.40.0.12/24。

然后配置路由：

```
ip netns exec test.0 ip route add default via 10.40.0.1
```

查看：

```shell
root@localhost:~/k3s# ip netns exec test.0 ip route
default via 10.40.0.1 dev veth1 linkdown 
10.40.0.0/24 dev veth1 proto kernel scope link src 10.40.0.10 linkdown 
```

同理，配置`test.1`的路由。

这时候去ping另外一个容器网络空间，会发现不通。

```shell
root@localhost:~/k3s# ip netns exec test.0 ping 10.40.0.12
PING 10.40.0.12 (10.40.0.12) 56(84) bytes of data.
```

6. 创建网桥，使不同容器网络空间可以通信

```
ip link add br0 type bridge
```

将veth0和veth2加入到网桥：

```
ip link set veth0 master br0
ip link set veth2 master br0
```

查看：
```shell
root@localhost:~/k3s# brctl show
bridge name	   bridge id		    STP enabled	  interfaces
br0		       8000.5a483fbf11ee	no		       veth0
							                       veth2
```

再次ping两个不同容器网络空间，发现依然不通。 原因在于网桥无法转接。

7. 配置网桥的IP地址

将网关路由IP配置给网桥br0 

```
ip addr add 10.40.0.1/24 dev br0
```

8. 测试

```shell
root@localhost:~/k3s# ip netns exec test.0 ping 10.40.0.12
PING 10.40.0.12 (10.40.0.12) 56(84) bytes of data.
64 bytes from 10.40.0.12: icmp_seq=1 ttl=64 time=0.050 ms
64 bytes from 10.40.0.12: icmp_seq=2 ttl=64 time=0.109 ms
64 bytes from 10.40.0.12: icmp_seq=3 ttl=64 time=0.108 ms
64 bytes from 10.40.0.12: icmp_seq=4 ttl=64 time=0.101 ms
64 bytes from 10.40.0.12: icmp_seq=5 ttl=64 time=0.108 ms
64 bytes from 10.40.0.12: icmp_seq=6 ttl=64 time=0.107 ms
```

当时ping本容器网络的IP时，则出现ping不通：

```shell
root@localhost:~/k3s# ip netns exec test.0 ping 10.40.0.10
PING 10.40.0.10 (10.40.0.10) 56(84) bytes of data.

```

原因是lo网络设备没开启：
```
root@localhost:~/k3s# ip netns exec test.0 ip addr
1: lo: <LOOPBACK> mtu 65536 qdisc noop state DOWN group default qlen 1000
    link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00
22: veth1@if23: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc noqueue state UP group default qlen 1000
    link/ether b2:f4:f2:2d:1d:63 brd ff:ff:ff:ff:ff:ff link-netnsid 0
    inet 10.40.0.10/24 scope global veth1
       valid_lft forever preferred_lft forever
    inet6 fe80::b0f4:f2ff:fe2d:1d63/64 scope link 
       valid_lft forever preferred_lft forever
```

开启lo设备：

```
ip netns exec test.0 ip link set lo up
```

这时候ping就正常了

```shell
root@localhost:~/k3s# ip netns exec test.0 ping 10.40.0.10
PING 10.40.0.10 (10.40.0.10) 56(84) bytes of data.
64 bytes from 10.40.0.10: icmp_seq=1 ttl=64 time=0.017 ms
64 bytes from 10.40.0.10: icmp_seq=2 ttl=64 time=0.054 ms
64 bytes from 10.40.0.10: icmp_seq=3 ttl=64 time=0.031 ms
64 bytes from 10.40.0.10: icmp_seq=4 ttl=64 time=0.053 ms
64 bytes from 10.40.0.10: icmp_seq=5 ttl=64 time=0.021 ms
```

9. 设置路由规则，使其能够跨主机通信

