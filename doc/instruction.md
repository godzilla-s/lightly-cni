# 如何开发一款kubernetes CNI插件

**目标:**

1. 了解CNI的工作机制
2. 掌握CNI插件的开发与部署

## 背景知识

### 什么时CNI

**历史背景**

CNI的发展历程反映了容器网络技术的演进：

+ ​​早期混乱期​​：在容器技术早期，每个容器平台(如Docker、Kubernetes、Mesos)都有自己的网络实现方式，导致网络解决方案需要为每个平台单独适配，造成大量重复工作。  

+ ​CNI的诞生​​：CNI最初由CoreOS在2015年提出，目的是为容器网络提供一个通用的接口标准。它从rkt的网络模型中发展而来，后来得到了包括Kubernetes在内的多个容器平台的支持。

+ 与CNM的竞争​​：Docker提出了自己的网络模型CNM(Container Network Model)，但CNM设计更为复杂且专为Docker定制，难以支持其他容器运行时。相比之下，CNI更加轻量和通用，最终被Kubernetes采纳为标准。

+ ​Kubernetes的采用​​：随着Kubernetes的崛起，CNI因其简单性和灵活性成为Kubernetes事实上的网络标准。Kubernetes通过CNI接口将网络功能外包给专门的网络插件。

+ ​​持续演进​​：CNI规范经过多次迭代，从最初的简单功能发展到支持更复杂的网络场景，如网络策略、服务发现等。

+ ​​生态系统繁荣​​：随着CNI被广泛接受，出现了大量CNI插件实现，如Calico、Flannel、Weave、Cilium等，每种插件针对不同的使用场景进行了优化


### CNI的运行机制

## 技术原理

### 基础原理

+ 网络模型（二层网络、三层网络）
+ kubernetes网络模型
+ 基础概念（AI整理报告）

### Pod内部网络

kubernetes是由一个个Pod组成的服务编排系统，每个Pod相互独立，如何保证Pod的独立性： 

linux namespace -> net namespace

+ Pod是如何创建网络
+ Pod内部网络的特点

### 同节点上Pod的通信

### 不同节点Pod的通信

## 开发与部署

### CNI介绍

### 开发准备与流程

### 配置与部署

## 总结

### 当前主流的CNI插件技术风格