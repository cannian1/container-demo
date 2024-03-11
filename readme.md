# 容器原理

## namespace（命名空间）

只能看到进程ID的命名空间、它自己的主机名和它的命名空间等，不影响宿主机

使用 syscall 设置命名空间

```go
const (
CLONE_VM             = 0x00000100 // set if VM shared between processes
CLONE_FS             = 0x00000200 // set if fs info shared between processes
CLONE_FILES          = 0x00000400 // set if open files shared between processes
CLONE_SIGHAND        = 0x00000800 // set if signal handlers and blocked signals shared
CLONE_PIDFD          = 0x00001000 // set if a pidfd should be placed in parent
CLONE_PTRACE         = 0x00002000 // set if we want to let tracing continue on the child too
CLONE_VFORK          = 0x00004000 // set if the parent wants the child to wake it up on mm_release
CLONE_PARENT         = 0x00008000 // set if we want to have the same parent as the cloner
CLONE_THREAD         = 0x00010000 // Same thread group?
CLONE_NEWNS          = 0x00020000 // New mount namespace group
CLONE_SYSVSEM        = 0x00040000 // share system V SEM_UNDO semantics
CLONE_SETTLS         = 0x00080000 // create a new TLS for the child
CLONE_PARENT_SETTID  = 0x00100000 // set the TID in the parent
CLONE_CHILD_CLEARTID = 0x00200000 // clear the TID in the child
CLONE_DETACHED       = 0x00400000 // Unused, ignored
CLONE_UNTRACED       = 0x00800000 // set if the tracing process can't force CLONE_PTRACE on this clone
CLONE_CHILD_SETTID   = 0x01000000 // set the TID in the child
CLONE_NEWCGROUP      = 0x02000000 // New cgroup namespace
CLONE_NEWUTS         = 0x04000000 // New utsname namespace
CLONE_NEWIPC         = 0x08000000 // New ipc namespace
CLONE_NEWUSER        = 0x10000000 // New user namespace
CLONE_NEWPID         = 0x20000000 // New pid namespace
CLONE_NEWNET         = 0x40000000 // New network namespace
CLONE_IO             = 0x80000000 // Clone io context

// Flags for the clone3() syscall.

CLONE_CLEAR_SIGHAND = 0x100000000 // Clear any signal handler and reset to SIG_DFL.
CLONE_INTO_CGROUP   = 0x200000000 // Clone into a specific cgroup given the right permissions.

// Cloning flags intersect with CSIGNAL so can be used with unshare and clone3
// syscalls only:

CLONE_NEWTIME = 0x00000080 // New time namespace
)
```

## rootfs（根文件系统）

在Linux系统中，rootfs（根文件系统）是指一个进程可见的文件系统层次结构的根目录。它为进程提供了一个独立的文件系统视图，包括目录、设备文件、socket文件等。rootfs 可以是实际磁盘上的一个目录，也可以是通过各种机制（如aufs、overlayfs、union mounts等）联合挂载的多个目录的统一视图。

在容器技术中，rootfs 特别重要，因为它为容器提供了其运行所需的所有文件和目录。容器的rootfs通常包含了一个基本的操作系统文件和目录结构，以及应用程序所需的任何文件。当容器启动时，它的rootfs会被挂载为只读模式，并在其上层创建一个读写层（通常使用overlayfs或其他联合文件系统实现），这样容器就可以在其rootfs上进行写操作，而不会影响到底层的只读rootfs。

在Docker等容器运行时中，当你构建一个镜像时，实际上是在构建一个rootfs。当你运行一个容器时，Docker会基于镜像的rootfs创建一个容器实例，并在其上添加一个读写层。

### chroot vs pivot_root vs switch_root

`chroot` 只改变当前进程的“/”

`pivot_root` 改变当前 mount namespace 的“/”

`switch_root`和 `chroot` 类似，但是专门用来初始化系统时候使用的（initramfs），不仅会chroot，而且会删除旧根下的所有内容，释放内存，只能由pid=1的进程使用，其他地方用不到

pivot_root 隔离得更彻底

## Cgroup

Cgroup（Control Groups）是Linux内核的一个特性，用于限制、计算和隔离一组进程的资源使用（如CPU、内存、磁盘I/O等）。它允许系统管理员和程序更好地控制如何以及在何处分配系统资源。

Cgroup的主要功能包括：

1. 资源限制：您可以设置特定cgroup中进程可以使用的最大资源量，例如最大内存使用量或最大CPU时间。

2. 优先级分配：通过设置权重，您可以控制不同cgroup中的进程之间的资源分配优先级。

3. 资源统计：您可以监控cgroup中进程的资源使用情况，例如CPU使用时间或内存使用量。

4. 进程控制：您可以创建和销毁cgroup，并在它们之间移动进程。

Cgroup常用于容器和虚拟化技术中，以确保系统资源得到有效和公平的分配。

## 总结

**namespace 限制了容器能看到什么，cgroup 限制了容器能使用什么资源，rootfs 限制了容器在哪工作。**