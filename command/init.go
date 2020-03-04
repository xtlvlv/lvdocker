package command

import (
	"log"
	"os"
	"syscall"
)

/*
Init 初始化容器,主要是挂载文件系统,然后运行cmd,替换当前进程为要执行的程序进程
*/
func Init(command string) {
	log.Println("command:", command)

	// TODO: 注意这里
	// https://github.com/xianlubird/mydocker/issues/41#issuecomment-478799767
	// systemd 加入linux之后, mount namespace 就变成 shared by default, 所以你必须显示
	//声明你要这个新的mount namespace独立。
	// syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, "")
	if err := syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, ""); err != nil {
		log.Fatal("init.go22,", err)
		return
	}
	// MS_NOEXEC 本文件系统不允许执行其他程序
	// MS_NOSUID 不允许 set-user-ID 和 set-group-ID
	// MS_NODEV  默认参数
	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	if err := syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), ""); err != nil {
		log.Fatal("init.go 444 ", err)
		return
	}
	// cmd:=exec.Command(command)
	// cmd.Stdin=os.Stdin
	// cmd.Stdout=os.Stdout
	// cmd.Stderr=os.Stderr
	// if err=cmd.Run();err!=nil{
	// 	log.Fatal("init.go1",err)
	// }
	argv := []string{command}
	if err := syscall.Exec(command, argv, os.Environ()); err != nil {
		log.Fatal("init.go333 ", err.Error())
	}
}
