package command

import (
	"log"
	"modfinal/cgroups/subsystems"
	"os"
	"os/exec"
)

/*
创建Init程序工作目录
*/
func NewWorkDir(rootPath,containerName,imageName,volume string) error {
	CreateContainerLayer(rootPath,containerName)
	CreateMntPoint(rootPath,containerName)
	SetMountPoint(rootPath,containerName,imageName)
	CreateVolume(rootPath,volume,containerName)
	return nil
}

/*
生成rootPath/writerLayer文件夹
*/
func CreateContainerLayer(rootPath,containerName string) error {
	writerLayer:=rootPath+"/writerLayer/"+containerName
	if err:=os.MkdirAll(writerLayer,0777);err!=nil{
		log.Fatal("aufs.go writerLayer ERROR,",err)
	}
	log.Println("创建可写层:",writerLayer)
	return nil
}

/*
生成mnt文件夹
*/
func CreateMntPoint(rootPath,containerName string) error {
	mnt:=rootPath+"/mnt/"+containerName
	if err:=os.MkdirAll(mnt,0777);err!=nil{
		log.Fatal("aufs.go mnt ERROR,",err)
	}
	log.Println("创建临时挂载点:",mnt)
	return nil
}

/*
挂载aufs文件系统
如果指定imageName,就在rootPath/images目录中找这个镜像,找到后解压到images中,用这个目录作为基础镜像
如果找不到就报错返回
*/
func SetMountPoint(rootPath,containerName,imageName string) error {
	imageDir:=""
	if imageName==""{
		imageDir = "/busybox"
	}else{
		imageDir=ImageDir(rootPath,imageName)
	}
	//dirs :="dirs="+rootPath+"/writerLayer/"+containerName+":"+rootPath+"/busybox"
	dirs :="dirs="+rootPath+"/writerLayer/"+containerName+":"+rootPath+imageDir
	mnt := rootPath+"/mnt/"+containerName
	if _,err:=exec.Command("mount","-t","aufs","-o",dirs,"none",mnt).CombinedOutput();err!=nil{
		log.Fatal("aufs.go mount aufs ERROR,",err)
	}
	log.Println("成功将busybox与writerLayer使用AUFS挂载到mnt上")
	return nil
}

func ImageDir(rootPath,imageName string) string{
	imageTar:=rootPath+"/images/"+imageName+".tar"
	 exist,_:=subsystems.PathExists(imageTar)
	 if !exist{
	 	log.Fatalf("%v镜像不存在\n",imageName)
	 }
	imagePath:=rootPath+"/images/"+imageName
	if err:=os.Mkdir(imagePath,0777);err!=nil{
		log.Fatalf("创建%v文件夹失败\n",imagePath)
	}
	_,err:=exec.Command("tar","-xf",imageTar,"-C",imagePath).CombinedOutput()
	if err!=nil{
		log.Fatal("aufs.go 解压镜像失败,",err)
	}
	return "/images/"+imageName
}

/*
清理工作,删除创建的文件夹
*/
func ClearWorkDir(rootPath,containerName,imageName,volume string)  {
	ClearVolume(rootPath,containerName,volume)
	ClearMountPoint(rootPath,containerName)
	ClearWriterLayer(rootPath,containerName)
	ClearImageDir(rootPath,imageName)
}

func ClearImageDir(rootPath,imageName string){
	if imageName==""{
		return
	}
	imageDir:=rootPath+"/images/"+imageName
	if err:=os.RemoveAll(imageDir);err != nil {
		log.Fatal("aufs.go 删除镜像目录失败,",err)
	}
	log.Println("成功删除镜像目录:",imageDir)
}

/*
卸载挂载点,删除mnt目录
*/
func ClearMountPoint(rootPath,containerName string)  {
	mnt:=rootPath+"/mnt/"+containerName
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
		log.Fatal("aufs.go remove mnt ERROR,",err)
	}
	log.Println("成功删除mnt目录:",mnt)
}

/*
删除可写层
*/
func ClearWriterLayer(rootPath,containerName string)  {
	writerLayer:=rootPath+"/writerLayer/"+containerName
	if err:=os.RemoveAll(writerLayer);err != nil {
		log.Fatal("aufs.go 删除可写层失败,",err)
	}
	log.Println("成功删除可写层:",writerLayer)
}
