package command

import (
	"log"
	"os"
	"os/exec"
	"syscall"
)

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