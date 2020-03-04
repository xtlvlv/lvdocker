package command

import (
	"log"
	"os"
	"os/exec"
	"syscall"
)

const rootDir  = "/home/lvkou/E/Task/毕业设计/root"

/*
Run run调用函数
*/
func Run(command string, tty bool, memory string) {

	reader,writer,err:=os.Pipe()
	if err!=nil{
		log.Fatal("run.go os.Pipe() Error")
		return
	}
	// cmd := exec.Command(command)
	// cmd := exec.Command("/proc/self/exe", "init", command)
	//args := []string{"init", command}
	//cmd := exec.Command("/proc/self/exe", args...)

	// 使用管道给子进程传输命令,就不用参数了
	cmd:=exec.Command("/proc/self/exe","init")

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWPID |
			syscall.CLONE_NEWIPC |
			syscall.CLONE_NEWNET |
			syscall.CLONE_NEWUTS |
			syscall.CLONE_NEWNS,
	}
	// 改变程序运行目录,执行/bin/sh后,用ls就会看到rootDir目录中的内容
	// 后面可以把这个参数化,即用户指定执行目录
	// 但这个只是改变了工作目录,使用pwd还是相对系统的目录,还需要使用pivot_root将这个目录变为根目录
	cmd.Dir=rootDir+"/busybox"

	// 这个是为了把读端传送给子进程,子进程就能通过reader从管道中读出数据,也就是要运行的程序
	cmd.ExtraFiles=[]*os.File{reader}
	sendInitCommand(command,writer)

	if tty {
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
	}

	/* Start()非阻塞运行 */
	if err := cmd.Start(); err != nil {
		log.Fatal("run.go1", err)
	}
	//subsystems.Set(memory)
	//subsystems.Apply(strconv.Itoa(cmd.Process.Pid))
	//defer subsystems.Remove()

	cmd.Wait()
}

func sendInitCommand(command string,writer *os.File)  {
	_,err:=writer.Write([]byte(command))
	if err != nil{
		log.Fatal("run.go 写入管道失败")
		return
	}
	writer.Close()
}