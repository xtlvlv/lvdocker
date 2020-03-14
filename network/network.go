package network

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/tabwriter"
)
import "github.com/vishvananda/netlink"



var(
	defaultNetworkPath=" "
	drivers=map[string]NetworkDriver{}
	networks=map[string]*Network{}
)

/*
网络
是容器的一个集合,在这个网络上的容器可以通过这个网络互相通信
 */
type Network struct {
	Name string	// 网络名
	IpRange *net.IPNet	// 地址段
	DriverName string	// 网络驱动名
}

/*
网络端点
连接容器与网络
 */
type EndPoint struct {
	Id 			string				`json:"id"`
	Device 		netlink.Veth		`json:"dev"`
	IpAddress 	net.IP				`json:"ip"`
	MacAddress 	net.HardwareAddr	`json:"mac"`
	PortMapping []string			`json:"portmapping"`
	NetWork 	*Network


}

/*
网络驱动接口
不同的驱动对网络的创建,连接和销毁的策略不同,这是驱动接口
 */
type NetworkDriver interface {
	Name() string	// 驱动名
	Create(subnet string,name string) (*Network,error)	// 创建网络
	Delete(network Network) error	// 删除网络
	Connect(network *Network,endpoint *EndPoint) error	// 连接容器网咯端点到网络
	Disconnect(network Network,endpoint *EndPoint) error // 从网络上移除网络端点
}


/*
dump就是把当前网络信息保存到dumpPath路径上,load就是把dump路径上的信息加载到程序中
 */
func (nw *Network)dump(dumpPath string) error {
	if _,err:=os.Stat(dumpPath);err!=nil{
		if os.IsNotExist(err){
			os.MkdirAll(dumpPath,0644)
		}else{
			log.Fatal("dump() error,",err)
		}
	}
	nwPath:=path.Join(dumpPath,nw.Name)
	nwFile,err:=os.OpenFile(nwPath,os.O_TRUNC|os.O_WRONLY|os.O_CREATE,0644)
	if err!=nil{
		log.Fatal("dump() 打开文件失败,",err)
	}
	defer nwFile.Close()

	nwJson,err:=json.Marshal(nw)
	if err!=nil{
		log.Fatal("dump() json序列化失败,",err)
	}

	_,err=nwFile.Write(nwJson)
	if err!=nil{
		log.Fatal("dump() 写入json失败,",err)
	}
	return nil
}

func (nw *Network)load(dumpPath string) error {
	nwConfigFile,err:=os.Open(dumpPath)
	defer nwConfigFile.Close()
	if err!=nil{
		log.Fatal("load() 打开文件失败,",err)
	}
	nwJson:=make([]byte,2000)
	n,err:=nwConfigFile.Read(nwJson)
	if err!=nil{
		log.Fatal("load() 读取文件失败,",err)
	}
	err=json.Unmarshal(nwJson[:n],nw)
	if err!=nil{
		log.Fatal("load() 解析json失败,",err)
	}
	return nil
}

/*
初始化方法
1. 初始化一个网桥驱动BridgeNetworkDriver并放到drivers中
2. 如果默认路径不存在则创建
3. 加载所有现存网络并放到networks中
 */
func Init() error {
	// 加载网络驱动
	var bridgeDriver=BridgeNetworkDriver{}
	drivers[bridgeDriver.Name()]=&bridgeDriver
	// 判断网络的配置目录是否存在,不存在则创建
	if _,err:=os.Stat(defaultNetworkPath);err!=nil{
		if os.IsNotExist(err){
			os.MkdirAll(defaultNetworkPath,0644)
		}else{
			log.Fatal("Init() 创建文件夹失败 error,",err)
		}
	}
	// 检查网络配置目录中的所有文件,这个函数会遍历第一个参数目录中的所有文件,并用第二个参数中的函数处理每一个文件
	filepath.Walk(defaultNetworkPath, func(nwPath string, info os.FileInfo, err error) error {

		// 如果是目录就跳过
		if strings.HasSuffix(nwPath,"/"){
			return nil
		}
		_,nwName:=path.Split(nwPath)
		nw:=&Network{
			Name:       nwName,
			IpRange:    nil,
			DriverName: "",
		}
		nw.load(nwPath)
		networks[nwName]=nw
		return nil
	})

	return nil
}

/*
展示所有网络
 */
func ListNetwork()  {
	w:=tabwriter.NewWriter(os.Stdout,12,1,3,' ',0)
	fmt.Fprint(w,"name\tIpRange\tDriver\n")
	for _,nw:=range networks{
		fmt.Fprintf(w,"%s\t%s\t%s\n",
			nw.Name,
			nw.IpRange.String(),
			nw.DriverName,
			)
	}
	if err:=w.Flush();err!=nil{
		log.Fatal("ListNetwork() flush error,",err)
	}
}

func DeleteNetwork(networkName string) error {

	nw,ok:=networks[networkName]
	if !ok{
		log.Fatal("没有这个网络:",networkName)
	}
	// 调用IPAM实例ipAllocator释放网络网关的IP
	ipAllocator.Release(nw.IpRange,&nw.IpRange.IP)
	// 调用网络驱动删除网络创建的设备与配置
	drivers[nw.DriverName].Delete(*nw)
	// 从网络的配置目录中删除该网络对应的配置文件
	nw.remove(defaultNetworkPath)
	return nil
}

/*
从网络的配置目录中删除该网络对应的配置文件
 */
func (nw *Network) remove(dumpPath string) error {
	if _,err:=os.Stat(path.Join(dumpPath,nw.Name));err!=nil{
		if os.IsNotExist(err){
			return nil
		}else{
			log.Fatal("network.go remove() error,",err)
		}
	}
	if err:=os.Remove(path.Join(dumpPath,nw.Name));err!=nil{
		log.Fatal("network.go remove() error2,",err)
	}
	return nil
}

func Connect(networkName string,cinfo *)  {

}

