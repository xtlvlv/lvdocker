package command

import (
	"log"
	"os/exec"
)

/*
保存镜像
1. 找到容器信息
2. 打包mnt文件夹到指定目录
 */
func Commit(containerName,imageName string)  {
	containerInfo,_:=GetContainerInfo(containerName)
	mntPath:=containerInfo.RootPath+"/mnt/"+containerName
	imageTar:=containerInfo.RootPath+"/images/"+imageName+".tar"
	_,err:=exec.Command("tar","-czf",imageTar,"-C",mntPath,".").CombinedOutput()
	if err!=nil{
		log.Fatal("commit.go 打包镜像失败,",err)
	}
	log.Println("打包镜像成功,镜像:",imageTar)
}
