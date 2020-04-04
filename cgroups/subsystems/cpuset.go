package subsystems

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
)

// cpu使用限制
type CpuSubSystem struct {
}

func (s *CpuSubSystem) Set(res *ResourceConfig) error {
	if res.CpuLimit!=""{
		content:=res.CpuLimit
		absolutePath:=""
		if absolutePath=FindAbsolutePath("cpuset");absolutePath==""{
			log.Fatal("cpuset.go 路径出错,")
			return fmt.Errorf("ERROR:absolutePath is empty!\n")
		}
		err := ioutil.WriteFile(path.Join(absolutePath,"cpuset.cpus"),[]byte(content),0777)
		if err!=nil{
			log.Fatal("cpuset.go 写入文件出错,",err)
			return fmt.Errorf("ERROR:写入文件出错!\n")
		}
		log.Println("限制cpuset成功:",content)
	}

	return nil
}

func (s *CpuSubSystem) Remove() error {
	absolutePath:=""
	absolutePath=FindAbsolutePath("cpuset")
	if absolutePath==""{
		log.Fatal("cpuset.go 路径寻找错误")
	}
	if err:=os.RemoveAll(absolutePath);err!=nil{
		log.Fatal("cpuset.go 删除文件夹失败",err)
	}
	return nil
}

func (s *CpuSubSystem) Apply(pid string) error {
	absolutePath:=""
	absolutePath=FindAbsolutePath("cpuset")
	if absolutePath==""{
		log.Fatal("cpuset.go apply 路径出错,")
		return fmt.Errorf("ERROR:apply absolutePath is empty!\n")
	}
	err := ioutil.WriteFile(path.Join(absolutePath,"tasks"),[]byte(pid),0644)
	if err!=nil{
		log.Fatal("cpuset.go 写入pid出错,")
		return fmt.Errorf("ERROR:写入pid出错!\n")
	}
	log.Println("把当前进程加入cpuset cgroup,当前进程:",pid)
	return nil
}

func (s *CpuSubSystem) Name() string {
	return "cpuset"
}

