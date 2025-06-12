
## 网关（gateway）

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

网桥的创建:
```shell
ip link add name <bridge_name> type bridge

# 查看
ip link show type bridge
```

### promisc模式（混杂模式）

混杂模式(Promiscuous  mode)，简称 Promisc  mode，俗称监听模式。

混杂模式通常被网络管理员用来诊断网络问题，但也会被无认证的、想偷听网络通信的人利用。根据维基百科的定义，混杂模式是指一个网卡会把它接收的所有网络流量都交给CPU，而不是只把它想转交的部分交给CPU。在`IEEE 802`定的网络规范中，每个网络帧都有一个目的MAC地址。在非混杂模式下，网卡只会接收目的MAC地址是它自己的单播帧，以及多播及广播帧；在混杂模式下，网卡会接收经过它的所有帧！

