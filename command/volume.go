package command

import (
	"log"
	"modfinal/cgroups/subsystems"
	"os"
	"os/exec"
	"strings"
)

/*
根据命令行的 -v /root/hostVolume:/myVolume 命令把hostVolume挂载到myVolume上
如果文件夹不存在,要创建
 */
func CreateVolume(rootPath,volume string) error {

	if volume!=""{
		containerMntPath := rootPath+"/mnt"
		// 宿主机路径要绝对路径或相对路径
		hostPath := strings.Split(volume,":")[0]
		exist,_:=subsystems.PathExists(hostPath)
		if !exist{
			if err:=os.Mkdir(hostPath,0777);err!=nil{
				log.Fatal("创建hostVolume ERROR,",err)
			}
		}
		mountPath:=strings.Split(volume,":")[1]
		containerPath := containerMntPath+mountPath
		if err:=os.Mkdir(containerPath,0777);err!=nil{
			log.Fatal("创建containerVolume ERROR",err)
		}
		dirs:="dirs="+hostPath
		if _,err:=exec.Command("mount","-t","aufs","-o",dirs,"none",containerPath).CombinedOutput();err!=nil{
			log.Fatal("volume 挂载AUFS ERROR,",err)
		}
		log.Println("数据卷volume挂载成功")
		log.Println("host dir:",hostPath)
		log.Println("container dir:",mountPath)
	}
	return nil
}

/*
卸载volume,并删除相应文件夹
 */
func ClearVolume(rootPath,volume string){
	if volume!=""{
		containerMntPath:=rootPath+"/mnt"
		mountPath:=strings.Split(volume,":")[1]
		containerPath:=containerMntPath+mountPath
		if _,err:=exec.Command("umount","-f",containerPath).CombinedOutput();err!=nil{
			log.Fatal("Volume umount ERROR,",err)
		}
		if err:=os.RemoveAll(containerPath);err!=nil{
			log.Fatal("删除containerVolume失败")
		}
		log.Println("container volume 删除成功")
	}

}
