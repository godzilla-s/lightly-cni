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

1. 创建

+ 创建Pod时，Kubernetes会首先创建一个Pause容器。
+ pause容器初始化Pod的共享命名空间，并为Pod中的其他容器提供网络和IPC命名空间的共享。
+ 其他应用容器加入pause容器所创建的命名空间

2. 命名空间共享

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