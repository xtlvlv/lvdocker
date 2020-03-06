package command

import (

	"log"

)

func Remove(containerName string)  {
	containerInfo,_:=GetContainerInfo(containerName)

	if containerInfo.Status!=STOP{
		Stop(containerName,false)	//先停止再删除
	}
	//RemoveContainerInfo(containerInfo)
	ClearContainerInfo(containerInfo.Name)
	ClearWorkDir(containerInfo.RootPath,containerName,containerInfo.Volume)
	log.Println("成功删除容器:",containerInfo.Name)
}

//func RemoveContainerInfo(containerInfo *ContainerInfo)  {
//	containerDir:=fmt.Sprintf(INFOLOCATION,containerInfo.Name)
//	if err:=os.RemoveAll(containerDir);err!=nil{
//		log.Fatal("remove.go 删除容器失败",err)
//	}
//
//}
