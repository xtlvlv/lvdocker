package command

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"syscall"
)

/*
停止容器
1. 找到进程pid杀掉进程
2. 改变配置文件中容器的状态为stop
 */
func Stop(containerName string)  {
	containerInfo,_:=GetContainerInfo(containerName)

	if containerInfo.Pid==""{
		log.Println("container not exist!")
		return
	}
	pid,err:=strconv.Atoi(containerInfo.Pid)
	if err!=nil{
		log.Fatal("stop.go ",err)
	}
	if err:=syscall.Kill(pid,syscall.SIGTERM);err!=nil{
		log.Fatal("stop.go kill ERROR,",err)
	}
	containerInfo.Status=STOP
	containerInfo.Pid=""
	UpdateContainerInfo(containerInfo)
	log.Println("成功停止容器")
}

func UpdateContainerInfo(containerInfo *ContainerInfo){
	jsonInfo,_:=json.Marshal(containerInfo)
	location:=fmt.Sprintf(INFOLOCATION,containerInfo.Name)
	file:=location+"/"+CONFIGNAME
	if err:=ioutil.WriteFile(file,[]byte(jsonInfo),0622);err!=nil{
		log.Fatal("更新容器信息失败,",err)
	}
}