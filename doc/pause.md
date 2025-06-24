# Pause容器

Pause容器又叫infra容器，是Kubernetes中的一个特殊类型的容器。它不执行任何实际的任务或进程，其主要作用是为Pod中的其他容器提供共享的网络命名空间和IPC（Inter-Process Communication）命名空间。

在Kubernetes中， Pod是由 pause + 应用容器（application containers）组成， 

## Pause容器的功能

1. 网络命名空间持有者

+ Pause容器在Pod中充当网络命名空间的主要进程，它创建了一个网络命名空间，并在其中设置Pod的网络配置，如IP地址、网络接口和路由规则。
+ Pod中的其他容器可以与Pause容器共享网络命名空间，从而实现容器间的网络通信。

2. 共享存储卷的挂载与管理

+ Pause容器还负责挂载和管理Pod级别的共享存储卷，如emptyDir卷。
+ 其他应用容器可以通过挂载相同的共享存储卷，与Pause容器共享文件系统，实现文件共享和数据共享。

3. 


## pause容器的工作原理

1. Pod创建

+ 创建Pod时，Kubernetes会首先创建一个Pause容器。
+ pause容器初始化Pod的共享命名空间，并为Pod中的其他容器提供网络和IPC命名空间的共享。
+ 其他应用容器加入pause容器所创建的命名空间
+ pause进程启动后会无线休眠，不执行任何业务逻辑

2. 命名空间共享

+ 网络命名空间： 所有容器共享同一个网络命名空间，这意味着Pod中的所有容器都可以通过localhost进行通信。
+ IPC命名空间： 所有容器可以通过系统V信号量、消息队列和共享内存进行通信。
+ PID命名空间： 容器内都可以看到彼此的进程

3. Pod终止

+ 当Pod被删除时，kubelet先终止pause容器，触发共享命名空间的释放（销毁）
+ 所关联的应用容器会被自动终止

## Pause容器的职责

pause容器承担两个重要的职责：

+ pause容器是Pod中Linux命名空间共享的基石

在Linux中，子进程会继承其父进程的命名空间。当然，你也可以通过unshare工具来创建一个拥有新的、独立的命名空间的进程。如果一个进程A处于运行状态，那么就可以将别的进程加入到A的命名空间中，这样就形成了一个Pod。Linux中通过系统调用setns就可以实现将一个进程加入到一个已经存在的命名空间的功能。Pod中的多个容器之间会共享命名空间，Docker的出现够使得Pod中的多个容器之间共享命名空间更容易一些：

  首先，通过docker启动一个pause容器。我们都知道容器的隔离是通过Linux内核的namespace命名空间机制来实现的，只要一个进程一直存在，它就会一直拥有一个命名空间。所以当pause容器就绪后，我们就得到了一个网络命名空间。我们可以在这个网络命名空间中加入新的容器来构建一个关联紧密的容器组。接着，我们可以继续启动新的容器，需要注意的是，在这个过程中我们始终将新启动的容器加入到pause容器的命名空间中。当上述这个过程完成后，一个Pod就形成了，这个Pod内的多个容器之间会共享同一个网络命名空间，而这个命名空间是属于pause容器的。这些容器之间可以直接通过localhost和端口来相互访问。

+ 当Pod内PID命名空间共享开启后，pause容器的地位等同于PID为1的进程，同时还负责Pod中僵死进程的回收工作

Linux中处于一个PID命名空间下的所有进程会形成一个树形结构，除了PID=1的”init“进程之外，每个进程都会有一个父进程。每个进程可以通过系统调用folk和exec来创建一个新的进程，而调用folk和exec的进程就会成为新创建进程的父进程。每个进程在操作系统的进程表中都会有一个记录，这条记录包含了对应的进程当前的状态和返回码等诸多信息。当一个进程结束其生命周期后，如果它的父进程没有通过系统调用wait获取返回码释放子进程的资源，那么这个进程就会变成僵死进程（zombie processes）。僵死进程不再运行，但是进程表中的记录仍然存在，除非父进程完成了子进程的回收工作。一般情况下，僵死进程的存在时间会非常短，但是世事无绝对，又或者如果僵死进程的数量太大，还是会浪费相当可观的系统资源的。僵尸进程一般由其父进程完成回收工作，但是如果父进程在子进程结束之前就已经结束运行的话，那么操作系统就会指派PID=1的”init“进程成为子进程新的父进程，由”init“进程调用wait获取子进程的返回码并完成子进程的资源回收和善后工作。

  在Docker中，每个容器都有其自己的PID命名空间，同时又可以将一个容器A加入到另一个容器B的命名空间中。此时，B就承担了”init“进程的职责，而诸如A等被加入到B的命名空间的容器就是”init“进程的“子进程”。假设一个Pod中没有pause容器只有业务容器，那么第一个创建的容器就需要承担起”init“进程的责任—维护命名空间、回收僵死进程等。但是业务容器可能并没有这样的能力来完成这些任务。所以kubernetes在Pod创建过程中会第一个将pause容器创建出来，整个Pod的网络命名空间就是pause容器的命名空间，pause容器会肩负起”init“进程（PID=1）的职责，同时完成僵死进程的回收工作。这就是为什么Pod中会有一个pause容器的原因。

  ”init“进程和僵死进程的概念只有在Pod内PID命名空间共享开启后才有意义，如果PID命名空间未共享，鉴于Pod中的每个容器都以PID=1的”init“进程来启动，所以它需要自己处理僵死进程。尽管应用通常不会大量产生子进程，所以一般情况下不会有什么问题，但是僵死进程耗尽内存资源确实是一个容易忽视的问题。此外，开启PID命名空间共享也能够保证在同一个Pod的多个容器之间相互发送信号来进行通信，所以Kubernetes集群默认开启PID命名空间共享真的是非常有益的。

## Pause 维护
pause容器主要负责维护Pod的网络命名空间和IPC命名空间，它需要一直运行以保持这些共享资源的可用性。
1. 查看pause容器

## pause容器的源码分析

pause容器的源码非常简单，代码也很精简, 其核心代码如下：

```c
#include <signal.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/types.h>
#include <sys/wait.h>
#include <unistd.h>

#define STRINGIFY(x) #x
#define VERSION_STRING(x) STRINGIFY(x)

#ifndef VERSION
#define VERSION HEAD
#endif

static void sigdown(int signo) {
    psignal(signo, "Shutting down, got signal");
    exit(0);
}

// 回收僵尸进程（SIGCHLD 信号触发）
static void sigreap(int signo) {
    while(waitpid(-1, NULL, WNOHANG) > 0)
        ;
}

int main(int argc, char **argv) {
  int i;
  for (i = 1; i < argc; ++i) {
    if (!strcasecmp(argv[i], "-v")) {
      printf("pause.c %s\n", VERSION_STRING(VERSION));
      return 0;
    }
  }

  if (getpid() != 1)
    /* Not an error because pause sees use outside of infra containers. */
    fprintf(stderr, "Warning: pause should be the first process\n");

  if (sigaction(SIGINT, &(struct sigaction){.sa_handler = sigdown}, NULL) < 0)
    return 1;
  if (sigaction(SIGTERM, &(struct sigaction){.sa_handler = sigdown}, NULL) < 0)
    return 2;
  if (sigaction(SIGCHLD, &(struct sigaction){.sa_handler = sigreap,
                                             .sa_flags = SA_NOCLDSTOP},
                NULL) < 0)
    return 3;

  for (;;)
    pause();
  fprintf(stderr, "Error: infinite loop terminated\n");
  return 42;
}
```

## 常见的问题

1. 能否不使用pause容器

不能，kubernetes依赖pause容器来维护pod的网络命名空间和IPC命名空间，生命周期管理，若直接删除pause容器，pod内的其他容器将无法正常工作。

2. pause占用资源情况怎样？

正常情况下，pause容器占用资源极少，除非被攻击。

3. 日和使用自定义的pause容器

修改kubelet的启动参数（注意自定义的镜像需要兼容kubernetes的CRI接口）： 
```
--pod-infra-container-image=<自定义pause容器>
```

4. pause容器是否一直处于运行状态？

是的，pause容器需要保持运行状态来维持命名空间，知道Pod被显示删除

5. pause容器会回收僵尸进程吗?

会的，pause进程会主动回收子进程，即使应用容器内的进程产生了僵尸进程，也会被pause容器清理。