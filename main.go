package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

// 复现 https://www.youtube.com/watch?v=8fi7uSYlOdc
// 这个视频中使用了 cgroup v1，但是 ubuntu 22.04 使用的是 cgroup v2，使用 ubuntu20.04.4 可以复现成功

// docker         run image <cmd> <params>
// go run main.go run       <cmd> <params>

func main() {
	switch os.Args[1] {
	case "run":
		run()
	case "child":
		child()
	default:
		log.Printf("error: unknown command %v\n", os.Args[1])
	}
}

func run() {
	fmt.Printf("Running %v as %d\n", os.Args[2:], os.Getpid())

	// 创建一个新的进程
	cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.SysProcAttr = &syscall.SysProcAttr{
		// 创建一个新的 UTS （UNIX Time-sharing System）命名空间，不同的 UTS 命名空间可以拥有不同的主机名和域名设置
		Cloneflags:   syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
		Unshareflags: syscall.CLONE_NEWNS,
	}

	// 如果不使用 /proc/self/exe，直接在这个函数调用 Sethostname 方法，那么
	// 在这一步才使用新的命名空间，因此上面设置 hostname 实际上影响了容器外的主机名
	err := cmd.Run()
	if err != nil {
		log.Printf("error to exec cmd: %v\n", err)
		return
	}
}

// 在使用了新的命名空间之后，这个函数调用 Sethostname 方法，设置容器的主机名
func child() {
	fmt.Printf("Running %v as %d\n", os.Args[2:], os.Getpid())
	cg()

	err := syscall.Sethostname([]byte("container"))
	if err != nil {
		log.Printf("unable to set hostname: %v\n", err)
		return
	}

	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// 需要准备 ubuntu base 文件系统
	// http://cdimage.ubuntu.com/ubuntu-base/releases
	// tar -zxvf 解压到 与本项目目录同级的 ubuntu-fs 目录
	// 执行 chroot 和 chdir 后，容器就不能访问宿主机的文件系统了
	err = syscall.Chroot("../ubuntu-fs")
	if err != nil {
		log.Printf("unable to chroot: %v\n", err)
		return
	}
	err = syscall.Chdir("../ubuntu-fs")
	if err != nil {
		log.Printf("unable to chdir: %v\n", err)
		return
	}

	err = syscall.Mount("proc", "proc", "proc", 0, "")
	if err != nil {
		log.Printf("unable to mount proc: %v\n", err)
		return
	}

	err = cmd.Run()
	if err != nil {
		log.Printf("error to exec cmd: %v\n", err)
		return
	}

	syscall.Unmount("proc", 0)
}

// 这是 cgroup v1 的写法，ubuntu 22.04 使用的是 cgroup v2，会报错
func cg() {
	cgroup := "/sys/fs/cgroup/"
	pids := filepath.Join(cgroup, "pids")
	err := os.Mkdir(filepath.Join(pids, "test"), 0755)
	if err != nil && !os.IsExist(err) {
		log.Printf("error to create cgroup: %v\n", err)
		return
	}
	must(os.WriteFile(filepath.Join(pids, "test/pids.max"), []byte("20"), 0700))
	// Remove the new cgroup in place after the container exits
	must(os.WriteFile(filepath.Join(pids, "test/notify_on_release"), []byte("1"), 0700))
	must(os.WriteFile(filepath.Join(pids, "test/cgroup.procs"), []byte(fmt.Sprintf("%d", os.Getpid())), 0700))
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
