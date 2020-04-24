package model

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"text/tabwriter"
	"time"
	"modfinal/cgroups/subsystems"
)

type ContainerInfo struct {
	Pid	string	`json:"pid"`					// 容器的init进程在宿主机的PID
	Id	string	`json:"id"`						// 容器Id
	Name 	string	`json:"name"`				// 容器名
	ImageName string `json:"image"`				// 镜像名
	Command 	string	`json:"command"`		// 容器内init运行命令
	CreateTime	string	`json:"createTime"`
	Status		string	`json:"status"`
	Volume		string	`json:"volume"`			// 容器的数据卷
	RootPath	string	`json:"rootPath"`		// 容器的操作目录
	PortMapping []string `json:"portmapping"`	// 端口映射
}

var (
	RUNNING		string = "running"
	STOP		string = "stoped"
	//EXIT		string = "exited"
	CONTAINS	string = "/home/lvkou/E/Task/毕业设计/root/containers"
	INFOLOCATION	string = "/home/lvkou/E/Task/毕业设计/root/containers/%s"	// 存储容器信息的文件,%s是容器名字
	CONFIGNAME	string = "config.json"
)

/*
生成容器唯一ID
 */
func ContainerUUID() string {
	//str:=time.Now().UnixNano()
	//containerId:=fmt.Sprintf("%d%d",str,int(math.Abs(float64(rand.Intn(10)))))
	//log.Println("生成containerId:",containerId)
	//return containerId

	letterBytes := "1234567890"
	rand.Seed(time.Now().UnixNano())
	b:=make([]byte,10)
	for i:=range b{
		b[i]=letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)

}
//
//func writeUUID(uuid string)  {
//	ioutil.WriteFile("uuid.txt",[]byte(uuid),0644)
//}
//
//func readUUID() string {
//	data,_:=ioutil.ReadFile("uuid.txt")
//	return string(data)
//}

/*
存储容器信息
 */
func RecordContainerInfo( pid,containerName,imageName,id,command,volume,rootPath string){

	var containerInfo *ContainerInfo
	containerInfo=&ContainerInfo{
		Pid:        pid,
		Id:         id,
		Name:       containerName,
		ImageName:imageName,
		Command:    command,
		CreateTime: time.Now().Format("2006-01-02 15:04:05"),
		Status:     RUNNING,
		Volume:		volume,
		RootPath:	rootPath,
	}

	jsonInfo,err:=json.Marshal(containerInfo)
	if err!=nil{
		log.Fatal("containerInfo.go json序列化失败",err)
	}
	log.Printf("容器信息 jsonInfo:%s\n",string(jsonInfo))
	location:=fmt.Sprintf(INFOLOCATION,containerName)
	file := location+"/"+CONFIGNAME
	exist,_:=subsystems.PathExists(location)
	// 创建日志文件在之前执行,可能会先创建目录
	if !exist{
		if err:=os.Mkdir(location,0777);err != nil{
			log.Fatal("containerInfo.go 创建容器信息目录失败",err)
		}
	}
	if err:=ioutil.WriteFile(file,[]byte(jsonInfo),0777);err!=nil{
		log.Fatal("containerInfo.go 写入容器配置文件失败",err)
	}
}

/*
获取容器信息
 */
func GetContainerInfo(name string) (*ContainerInfo,error) {
	location:=fmt.Sprintf(INFOLOCATION,name)
	file:=location+"/"+CONFIGNAME
	containerInfo:=&ContainerInfo{}
	data,err:=ioutil.ReadFile(file)
	if err!=nil{
		log.Fatal("containerInfo.go 读信息失败,",err)
	}
	err =json.Unmarshal(data,containerInfo)
	if err!=nil{
		log.Fatal("containerInfo.go 解析json失败,",err)
	}
	return containerInfo,nil
}

/*
显示所有容器信息
 */
func ShowAllContainers()  {
	files,err:=ioutil.ReadDir(CONTAINS)
	if err!=nil{
		log.Fatal("containerInfo.go 读目录失败",err)
	}
	var containers []*ContainerInfo
	for _,file := range files{
		container,err:=GetContainerInfo(file.Name())
		if err!=nil{
			log.Fatal("containerInfo.go 读文件失败,",err)
		}
		containers=append(containers, container)
	}
	w:=tabwriter.NewWriter(os.Stdout,12,1,3,' ',0)
	fmt.Fprint(w,"Id\tName\tPid\tStatus\tCommand\tCreated\n")
	for _,item:=range containers{
		fmt.Fprintf(w,"%s\t%s\t%s\t%s\t%s\t%s\n",
			item.Id,
			item.Name,
			item.Pid,
			item.Status,
			item.Command,
			item.CreateTime)
	}
	if err:=w.Flush();err!=nil{
		log.Fatal("containerInfo.go 输出容器信息失败,",err)
	}
}

/*
容器退出时删除信息
1.获取这个容器名字
2.删除文件夹
 */
func ClearContainerInfo(name string)  {
	containerDir:=fmt.Sprintf(INFOLOCATION,name)
	if err:=os.RemoveAll(containerDir);err!=nil{
		log.Fatal("containerInfo.go 删除容器文件夹失败,",err)
	}
	log.Printf("成功删除 %s 容器信息\n",name)
}
