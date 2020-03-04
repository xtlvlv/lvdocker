package subsystems

import (
	"bufio"
	"io"
	"log"
	"os"
	"strings"
)

/*
FindGroupMountPoint 根据subsystem名字找到对应的hierarchy
从而可以在该hierarchy上创建子cgroup,再把进程加入到该cgroup限制中
*/
func FindGroupMountPoint(subsystem string) string {
	file, err := os.Open("/proc/self/mountinfo")
	if err != nil {
		log.Println("utils.go111,open error:", err)
		return ""
	}
	defer file.Close()
	bufRead := bufio.NewReader(file)
	for {
		line, err := bufRead.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				return ""
			}
		}
		/* 这个文件中的记录
		39 31 0:34 / /sys/fs/cgroup/memory rw,nosuid,nodev,noexec,relatime - cgroup cgroup rw,memory
		*/
		parts := strings.Split(string(line), " ")
		if strings.Contains(parts[len(parts)-1], subsystem) {
			return parts[4]
		}
	}
}

/*
FindAbsolutePath FindGrou***()只是找到了subsystem所在的目录,要自己创建cgroup,就是在该目录下自己创建一个目录
目录名字自己指定,这里先统一用lvdocker
1. 找到subsystem目录
2. 把自己的目录名字加到后面组成完整路径
3. 看该目录存在不,存在返回这个绝对路径,不存在就返回
*/
func FindAbsolutePath(subsystem string) string {
	path := FindGroupMountPoint(subsystem)
	if path != "" {
		absolutePath := path + "/" + CgroupDirName
		isExist, err := PathExists(absolutePath)
		if err != nil {
			log.Println("utils.go,路径不存在 ", err)
			return ""
		}
		if !isExist {
			err := os.Mkdir(absolutePath, os.ModePerm)
			if err != nil {
				log.Println("utils.go 创建文件夹失败 ", err)
				return ""
			}
		}

		return absolutePath
	}

	return ""
}

/*
PathExists 判断路径是否存在
*/
func PathExists(path string) (bool, error) {

	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}

	return true, err
}
