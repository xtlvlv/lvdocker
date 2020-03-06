package command

import (
	"log"
	"os"
	"os/exec"
	"strconv"
	"syscall"
)

const rootDir  = "/home/lvkou/E/Task/毕业设计/root"

/*
Run run调用函数
*/
func Run(command string, tty bool, memory,volume,containerName string) {

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
	log.Println("当前rootDir为:",rootDir)
	NewWorkDir(rootDir,volume)	// 这里如果出错会直接报错并停止
	cmd.Dir=rootDir+"/mnt"
	//defer ClearWorkDir(rootDir,volume)

	// 这个是为了把读端传送给子进程,子进程就能通过reader从管道中读出数据,也就是要运行的程序
	cmd.ExtraFiles=[]*os.File{reader}
	sendInitCommand(command,writer)

	id:=ContainerUUID()
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
	//subsystems.Set(memory)
	//subsystems.Apply(strconv.Itoa(cmd.Process.Pid))
	//defer subsystems.Remove()


	//RecordContainerInfo("测试",containerName,id,command)
	RecordContainerInfo(strconv.Itoa(cmd.Process.Pid),containerName,id,command,volume,rootDir)

	// 只有指定it的时候等待子进程结束,否则直接结束,子进程就由系统1进程管理
	if tty{
		cmd.Wait()
		ClearContainerInfo(containerName)
		ClearWorkDir(rootDir,volume)	// 如果后台运行的话,这文件夹就不删除了
	}

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

/*
创建Init程序工作目录
 */
func NewWorkDir(rootPath,volume string) error {
	CreateContainerLayer(rootPath)
	CreateMntPoint(rootPath)
	SetMountPoint(rootPath)
	CreateVolume(rootPath,volume)
	return nil
}

/*
生成rootPath/writerLayer文件夹
 */
func CreateContainerLayer(rootPath string) error {
	writerLayer:=rootPath+"/writerLayer"
	if err:=os.Mkdir(writerLayer,0777);err!=nil{
		log.Fatal("run.go writerLayer ERROR,",err)
	}
	log.Println("创建可写层:",writerLayer)
	return nil
}

/*
生成mnt文件夹
 */
func CreateMntPoint(rootPath string) error {
	mnt:=rootPath+"/mnt"
	if err:=os.Mkdir(mnt,0777);err!=nil{
		log.Fatal("run.go mnt ERROR,",err)
	}
	log.Println("创建临时挂载点:",mnt)
	return nil
}

/*
挂载aufs文件系统
 */
func SetMountPoint(rootPath string) error {
	dirs :="dirs="+rootPath+"/writerLayer:"+rootPath+"/busybox"
	mnt := rootPath+"/mnt"
	if _,err:=exec.Command("mount","-t","aufs","-o",dirs,"none",mnt).CombinedOutput();err!=nil{
		log.Fatal("run.go mount aufs ERROR,",err)
	}
	log.Println("成功将busybox与writerLayer使用AUFS挂载到mnt上")
	return nil
}

/*
清理工作,删除创建的文件夹
 */
func ClearWorkDir(rootPath,volume string)  {
	ClearVolume(rootPath,volume)
	ClearMountPoint(rootPath)
	ClearWriterLayer(rootPath)
}

/*
卸载挂载点,删除mnt目录
 */
func ClearMountPoint(rootPath string)  {
	mnt:=rootPath+"/mnt"
	//if _,err:=exec.Command("umount","-f",mnt).CombinedOutput();err!=nil{
	//	log.Println("umount path:",mnt)
	//	log.Fatal("run.go umount ERROR,",err)
	//}
	cmd := exec.Command("umount", mnt)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatal("error when umount mnt ", mnt, err)
	}
	log.Println("成功卸载mnt目录")
	if err:=os.RemoveAll(mnt);err!=nil{
		log.Fatal("run.go remove mnt ERROR,",err)
	}
	log.Println("成功删除mnt目录:",mnt)
}

/*
删除可写层
 */
func ClearWriterLayer(rootPath string)  {
	writerLayer:=rootPath+"/writerLayer"
	if err:=os.RemoveAll(writerLayer);err != nil {
		log.Fatal("run.go 删除可写层失败,",err)
	}
	log.Println("成功删除可写层:",writerLayer)
}















