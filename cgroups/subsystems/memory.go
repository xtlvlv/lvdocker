package subsystems

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
)

/*
设置内存的限制值
1.就是在对应的cgroup中的memory.limit_in_bytes写入特定值
*/
func Set(content string) error {
	absolutePath:=""
	if absolutePath=FindAbsolutePath("memory");absolutePath==""{
		log.Fatal("memeory.go 路径出错,")
		return fmt.Errorf("ERROR:absolutePath is empty!\n")
	}
	err := ioutil.WriteFile(path.Join(absolutePath,"memory.limit_in_bytes"),[]byte(content),0777)
	if err!=nil{
		log.Fatal("memeory.go 写入文件出错,",err)
		return fmt.Errorf("ERROR:写入文件出错!\n")
	}
	return nil
}

/*
对进程应用这个限制
1.把进程id加入到这个cgroup文件夹下的tasks文件中
*/
func Apply(pid string) error {
	absolutePath:=""
	absolutePath=FindAbsolutePath("memory")
	if absolutePath==""{
		log.Fatal("memeory.go apply 路径出错,")
		return fmt.Errorf("ERROR:apply absolutePath is empty!\n")
	}
	err := ioutil.WriteFile(path.Join(absolutePath,"tasks"),[]byte(pid),0644)
	if err!=nil{
		log.Fatal("memeory.go 写入pid出错,")
		return fmt.Errorf("ERROR:写入pid出错!\n")
	}
	return nil
}

/*
资源删除,在进程结束的时候把这个资源限制解除,其实就是把对应的文件夹删除
*/
func Remove()error{
	absolutePath:=""
	absolutePath=FindAbsolutePath("memory")
	if absolutePath==""{
		log.Fatal("memory.go 路径寻找错误")
	}
	if err:=os.RemoveAll(absolutePath);err!=nil{
		log.Fatal("memory.go 删除文件夹失败",err)
	}
	return nil
}
