package command

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

var (
	CONTAINERLOGS	=	"container.log"
)

func Logs(containerName string)  {
	data:=ReadLogs(containerName)
	fmt.Println(data)
}

/*
创建log文件
 */
func GetLogFile(containerName string) (*os.File) {
	path := fmt.Sprintf(INFOLOCATION,containerName)
	logPath:=path+"/"+CONTAINERLOGS
	if err:=os.Mkdir(path,0622);err!=nil{
		log.Fatal("log.go 创建log文件夹失败,",err)
	}
	if file,err:=os.Create(logPath);err!=nil{
		log.Fatal("log.go 创建log文件失败,",err)
	}else {
		return file
	}
	return nil
}

/*
读日志信息
 */
func ReadLogs(containerName string) string{

	path := fmt.Sprintf(INFOLOCATION,containerName)
	logPath:=path+"/"+CONTAINERLOGS
	data,err:=ioutil.ReadFile(logPath)
	if err!=nil{
		log.Fatal("log.go 读日志信息失败,",err)
	}
	return string(data)
}





