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
	// cmd := exec.Command(command)
	// cmd := exec.Command("/proc/self/exe", "init", command)
	args := []string{"init", command}
	cmd := exec.Command("/proc/self/exe", args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWPID |
			syscall.CLONE_NEWIPC |
			syscall.CLONE_NEWNET |
			syscall.CLONE_NEWUTS |
			syscall.CLONE_NEWNS,
	}

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
