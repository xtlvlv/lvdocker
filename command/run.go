package command

import (
	"log"
	"modfinal/model"
	"modfinal/network"
	"os"
	"os/exec"
	"strconv"
	"syscall"
)

const rootDir  = "/home/lvkou/E/Task/毕业设计/root"

/*
Run run调用函数
*/
func Run(command string, tty bool, memory,volume,containerName,nw string,portMapping []string) {

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
	// 后面可以把这个参数化,即用户指定执行目录,就是rootDir用户指定
	// 但这个只是改变了工作目录,使用pwd还是相对系统的目录,还需要使用pivot_root将这个目录变为根目录,这样init
	//cmd.Dir=rootDir+"/busybox"
	containerRootDir:=rootDir+"/mnt/"+containerName
	log.Println("当前rootDir为:",rootDir)
	NewWorkDir(rootDir,containerName,volume)	// 这里如果出错会直接报错并停止
	cmd.Dir=containerRootDir
	//defer ClearWorkDir(rootDir,volume)

	// 这个是为了把读端传送给子进程,子进程就能通过reader从管道中读出数据,也就是要运行的程序
	cmd.ExtraFiles=[]*os.File{reader}
	sendInitCommand(command,writer)

	id:= model.ContainerUUID()
	if containerName==""{
		containerName=id
	}

	if tty {
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
	}else {
		logFile:=GetLogFile(containerName)
		cmd.Stdout=logFile
		cmd.Stderr=logFile
	}

	/* Start()非阻塞运行 */
	if err := cmd.Start(); err != nil {
		log.Fatal("run.go1", err)
	}

	if nw!=""{
		network.Init()
		containerInfo:=&model.ContainerInfo{
			Pid:         strconv.Itoa(cmd.Process.Pid),
			Id:          id,
			Name:        containerName,
			PortMapping: portMapping,
		}
		network.Connect(nw,containerInfo)
	}

	//subsystems.Set(memory)
	//subsystems.Apply(strconv.Itoa(cmd.Process.Pid))
	//defer subsystems.Remove()


	//RecordContainerInfo("测试",containerName,id,command)
	model.RecordContainerInfo(strconv.Itoa(cmd.Process.Pid),containerName,id,command,volume,rootDir)

	// 只有指定it的时候等待子进程结束,否则直接结束,子进程就由系统1进程管理
	if tty{
		cmd.Wait()
		// 要主动把容器停止,但是不用kill命令,
		Stop(containerName,tty)

		// 用rm命令删除,退出的时候不直接删除
		//ClearContainerInfo(containerName)
		//ClearWorkDir(rootDir,volume)
	}

	// 后台运行的文件需要用rm命令删除
}

func sendInitCommand(command string,writer *os.File)  {
	_,err:=writer.Write([]byte(command))
	if err != nil{
		log.Fatal("run.go 写入管道失败")
		return
	}
	writer.Close()
	log.Println("成功将命令发送给init,cmd:",command)
}

/*
创建rootPath/busybox工作目录
将busybox.tar解压到这个目录
 */
func getRootPath(rootPath string) string{

	return ""
}















