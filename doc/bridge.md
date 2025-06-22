
# Bridge(桥接)

简单来说，桥接就是把一台机器上的额若干网络接口连接起来，

Bridge（桥）是 Linux 上用来做 TCP/IP 二层协议交换的设备，与现实世界中的交换机功能相似。Bridge 设备实例可以和 Linux 上其他网络设备实例连接，既 attach 一个从设备，类似于在现实世界中的交换机和一个用户终端之间连接一根网线。当有数据到达时，Bridge 会根据报文中的 MAC 信息进行广播、转发、丢弃处理。

## linux的桥接实现

Linux内核支持网口的桥接，但是与单纯的交换机不同，交换机只是二层网络设备，对于接收到的报文，要么转发要么丢弃，小型的交换机里面只需要一块交换芯片即可，并不需要CPU。而运行在linux内核的机器本身就是一台主机，有可能是网络报文的目的地，其收到报文除了转发和丢弃，还有可能被送到网络协议栈的上层（网络层），从而被自己消化掉。

## 网桥功能

1. Max学习： 学习MAC地址

2. 报文转发

## bridge的工作原理

**概念：** 

网关（Gateway）是连接不同网络的重要组件，负责在不同网络之间转发数据包。网关不仅有转发数据的功能，还有地址转换、安全控制、数据监控等功能。在局域网中，通常使用路由器作为网关

**大致类型**

+ 默认网关： 用于所有不在本地网络中的目的地。
+ 特定网关： 用于特定网络或者主机
+ 虚拟网关： 在虚拟机化环境中使用，用于连接虚拟机之间的网络

**使用**

查看网关:
```shell
ip route show

# 查看默认网关
ip route show default

# 或者使用route命令
route -n
```

网关的操作（增删改）：

```shell
# 删除网关
ip route del default

# 添加网关
ip route add default via <gateway_ip> dev <interface>
```

## 网桥(bridge)

桥是一种网络设备，用于连接两个网络，使它们能够互相通信，同时，桥还能将数据在网段之间传递。桥接可以实现MAC地址绑定，将同一个子网的不同计算机之间的通信放到网卡层处理，减轻网络的负担。

是否可以理解为虚拟交换机?

网桥的创建:
```shell
ip link add name <bridge_name> type bridge

# 查看
ip link show type bridge
```

## 路由

路由（Routing）是网络通信的核心机制，负责​​决定数据包从源设备到目标设备的传输路径​​。它类似于现实中的“导航系统”，通过查询​​路由表​​选择最优路径，确保数据包正确到达目的地。

### 路由表

路由表(Routing Table)是网络设备（如路由器、服务器）用于决定数据包转发路径的核心数据结构，它记录网络地址与下一跳（Next Hop）之间的映射关系，帮组路由器判断数据包应该发送到哪个接口或者相邻设备。每个网络设备都维护一张​​路由表​​，存储如何转发数据包的规则。

**路由表的基本组成**

路由表由多条路由条目组成

比如下面某个机器上路由信息：
路由表信息:
```shell
root@m5-stag-svr6:~$ route -n
Kernel IP routing table
Destination     Gateway         Genmask         Flags Metric Ref    Use Iface
0.0.0.0         172.16.5.254    0.0.0.0         UG    100    0        0 ens3
10.0.0.0        172.16.5.84     255.255.255.0   UG    0      0        0 ens3
10.0.1.0        172.16.5.85     255.255.255.0   UG    0      0        0 virbr0
172.17.0.0      0.0.0.0         255.255.0.0     U     0      0        0 docker0
192.168.122.0   0.0.0.0         255.255.255.0   U     0      0        0 virbr0
```

**路由表的查询与匹配**

当路由器收到一个数据包时，它会根据一下步骤查询路由表，并决定数据包的转发路径：

1. 提取目标IP地址： 从数据包的IP头部获取目标IP地址
2. 最长前缀匹配（Longest Prefix Match）:
    + 路由器将目标地址与路由表中的每个目标网络进行比较
    + 使用子网掩码计算目标IP是否属于某个网络段
    + 如果多个条目匹配，选择子网掩码最长（网络范围最具体的）的条目(比如: 192.168.1.0/24 比 192.168.0.0/16 更具体)
3. 确定下一跳:
    + 如果匹配到条目，数据包将转发到指定的下一跳地址或者直接通过指定接口发送
    + 如果没有匹配到条目，且存在默认路由（0.0.0.0/0）,数据包发往默认网关，否则，丢弃数据包并返回“不可达”消息

下面通过一个例子说明：


查看路由信息

```shell
ip route show
```
或者
```shell
route -n
```

路由决策模拟（测试数据包如何路由到目标 IP）： 

```shell
ip route get <pod ip>
```

比如:
```shell
[root@localhost]# ip route get 10.0.2.223
10.0.2.223 vis 172.16.5.86 dev ens3 src 172.16.5.83 uid 0
    cache
```



Flags：总共有多个旗标，代表的意义如下：                        

+ U (route is up)：该路由是启动的；                       
+ H (target is a host)：目标是一部主机 (IP) 而非网域；                       
+ G (use gateway)：需要透过外部的主机 (gateway) 来转递封包；                       
+ R (reinstate route for dynamic routing)：使用动态路由时，恢复路由资讯的旗标；                       
+ D (dynamically installed by daemon or redirect)：已经由服务或转 port 功能设定为动态路由                       
+ M (modified from routing daemon or redirect)：路由已经被修改了；              
+ ! (reject route)：这个路由将不会被接受(用来抵挡不安全的网域！)
+ A (installed by addrconf)
+ C (cache entry)

### promisc模式（混杂模式）

混杂模式(Promiscuous  mode)，简称 Promisc  mode，俗称监听模式。

混杂模式通常被网络管理员用来诊断网络问题，但也会被无认证的、想偷听网络通信的人利用。根据维基百科的定义，混杂模式是指一个网卡会把它接收的所有网络流量都交给CPU，而不是只把它想转交的部分交给CPU。在`IEEE 802`定的网络规范中，每个网络帧都有一个目的MAC地址。在非混杂模式下，网卡只会接收目的MAC地址是它自己的单播帧，以及多播及广播帧；在混杂模式下，网卡会接收经过它的所有帧！

## 技巧总结

1. 如何校验Pod网络连通性

首先需要知道Pod的进程ID：

```shell
# 通过容器获取
# 获取容器ID
ctrctl ps -a |grep <pod容器内运行的程序> 
crictl inspect <容器ID> |grep pid

# 或者直接在主机上通过ps查找
ps -ef |grep <pod容器内运行的程序名>
```

接着通过进程ID获取Pod的netns:
```shell
ip netns identify <pid>
```

最后通过netns进行一系列的网络测试:

```shell
ip netns exec <netns> ping <目标IP>

# 或者
ip netns exec <netns> curl <目标IP>:<端口号>

ip netns exec <netns> ip addr
```

查看某个网桥下已加入的虚拟网卡设备：

```
brctl show <网桥名>
```