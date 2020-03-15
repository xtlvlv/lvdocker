package network

import (
	"encoding/json"
	"fmt"
	"github.com/vishvananda/netns"
	"log"
	"modfinal/model"
	"net"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"text/tabwriter"
)
import "github.com/vishvananda/netlink"



var(
	defaultNetworkPath="/home/lvkou/E/Task/毕业设计/root/network"
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
	log.Println("成功把网络信息持久化到:",nwPath)
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
	log.Printf("成功把%s网络信息读入程序",dumpPath)
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
		//if strings.HasSuffix(nwPath,"/"){
		//	return nil
		//}
		if info.IsDir(){
			return nil
		}
		_,nwName:=path.Split(nwPath)
		nw:=&Network{
			Name:       nwName,
			IpRange:    nil,
			DriverName: "",
		}
		log.Printf("正在读取%s 文件,文件名:%s\n",nwPath,nwName)
		nw.load(nwPath)
		networks[nwName]=nw
		return nil
	})
	log.Println("成功Init网络信息,把所有网络信息加载到程序中")
	return nil
}

/*

 */
func CreateNetwork(driver,subnet,name string) error {

	_,cidr,_:=net.ParseCIDR(subnet)
	ip,err:=ipAllocator.Allocate(cidr)
	if err!=nil{
		log.Fatal("CreateNetwork()1,",err)
	}
	cidr.IP=ip
	nw,_:=drivers[driver].Create(cidr.String(),name)
	nw.dump(defaultNetworkPath)
	log.Printf("成功创建network,dirver:%s,subnet:%s,name:%s\n",driver,subnet,name)
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
	log.Printf("成功删除network,name:%s\n",networkName)
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
	log.Printf("成功删除网络配置文件:%s\n",dumpPath)
	return nil
}

func Connect(networkName string,cinfo *model.ContainerInfo) error {
	network,ok:=networks[networkName]
	if !ok{
		log.Fatal("No such network:",networkName)
	}
	// 从网络的IP段中,分配容器IP地址
	ip,_:=ipAllocator.Allocate(network.IpRange)

	// 创建网络端点,设置网络端点的IP,网络和端口映射信息,供下面的配置调用
	ep:=&EndPoint{
		Id:          fmt.Sprintf("%s-%s",cinfo.Id,networkName),
		IpAddress:   ip,
		PortMapping: cinfo.PortMapping,
		NetWork:     network,
	}

	// 调用网络对应的网络驱动挂载,配置网络端点
	drivers[network.DriverName].Connect(network,ep)

	// 到容器的namespace中配置容器网络,设备IP地址和路由信息
	configEndpointIpAddressAndRoute(ep,cinfo)

	// 配置端口映射信息,例如mydocker run -p 8080:80
	configPortMapping(ep,cinfo)

	log.Printf("成功Connect %s 到 %s 网络.\n",cinfo.Name,networkName)
	return nil
}

func Disconnect(networkName string,cinfo *model.ContainerInfo) error {
	return nil
}

/*
进入容器Net Namespace,使得容器网络端点的Veth容器端,以及后续的配置都在容器的Net Namespace中执行.
1. 将容器的网络端点加入到容器的网络空间中
2. 锁定当前程序所执行的线程,使当前线程进入到容器的网络空间
3. 返回值是一个函数指针,执行这个返回函数才会退出容器的网络空间,回归到宿主机的网络空间
 */

func enterContainerNetns(enLink *netlink.Link,cinfo *model.ContainerInfo) func() {

	f,err:=os.OpenFile(fmt.Sprintf("/proc/%s/ns/net",cinfo.Pid),os.O_RDONLY,0)
	if err!=nil{
		log.Fatal("enterContainerNetns() 打开文件失败,",err)
	}
	nsFD:=f.Fd()
	runtime.LockOSThread()

	// 修改veth peer, 将其移到容器的 net namespace中
	if err:=netlink.LinkSetNsFd(*enLink,int(nsFD));err!=nil{
		log.Fatal("enterContainerNetns() 进入net 失败,",err)
	}

	// 获取当前网络的net namespace,以便在从容器的net namespace中退出后回到原来的net namespace中
	origns,err:=netns.Get()
	if err!=nil{
		log.Fatal("enterContainerNetns() 获取当前net namespace失败,",err)
	}

	// 设置当前进程到容器网络的namespace,
	if err=netns.Set(netns.NsHandle(nsFD));err!=nil{
		log.Fatal("enterContainerNetns() 进入容器网络失败,",err)
	}
	// 并在函数执行完后回到原来的net namespace
	// 调用如下函数就能将程序恢复到原生的net namespace
	return func() {
		netns.Set(origns)	// 恢复到原来net
		origns.Close()		// 关闭namespace文件
		runtime.UnlockOSThread()	// 取消线程锁定
		f.Close()			// 关闭namespace文件
	}
}

/*
配置容器网络端点的地址和路由
 */
func configEndpointIpAddressAndRoute(ep *EndPoint,cinfo *model.ContainerInfo) error {
	// Veth 的另一端
	peerLink,err:=netlink.LinkByName(ep.Device.PeerName)
	if err!=nil{
		log.Fatal("configEndpointIpAddressAndRoute() peer link 出错,",err)
	}

	// 将容器的网络端点加入到容器的网络空间中,并使得这个函数下面的操作都在这个网络空间中进行
	// 执行完函数后,恢复为默认的网络空间
	defer enterContainerNetns(&peerLink,cinfo)()


	// 获取容器的IP地址及网段,用于配置容器内部接口地址
	// 比如容器ip是192.168.1.2,网段是192.168.1.0/24,那么传出的IP字符串就是192.168.1.2/24,用于容器内Veth端点配置
	interfaceIP:=*ep.NetWork.IpRange
	interfaceIP.IP=ep.IpAddress

	// 设置容器内Veth端点配置
	setInterfaceIp(ep.Device.PeerName,interfaceIP.String())
	// 启动容器内的Veth端点
	setInterfaceUp(ep.Device.PeerName)

	// net namespace 默认本地地址127.0.0.1的"lo"网卡是关闭状态的,启动它以保证容器访问自己的请求
	setInterfaceUp("lo")

	// 设置容器内的外部请求都通过容器内的Veth端点访问
	// 0.0.0.0/0的网段,表示所有的Ip地址段
	_,cidr,_:=net.ParseCIDR("0.0.0.0/0")

	// 构建要添加的路由数据,包括网络设备,网关IP及目的网段
	// 相当于route add -net 0.0.0.0/0 gw (Bridge网桥地址) dev (容器内的Veth端点设备)
	defaultRoute:=&netlink.Route{
		LinkIndex:  peerLink.Attrs().Index,
		Dst:        cidr,
		Gw:         ep.NetWork.IpRange.IP,
	}

	// 添加路由到容器的网络空间
	// RouteAdd() 相当于route add
	if err=netlink.RouteAdd(defaultRoute);err!=nil{
		log.Fatal("configEndpointIpAddressAndRoute() 添加路由失败,",err)
	}
	return nil
}

/*
配置宿主机到容器的端口映射,不然容器无法访问宿主机外部
 */
func configPortMapping(ep *EndPoint,cinfo *model.ContainerInfo) error {

	// 遍历容器端口映射列表
	for _,pm:=range ep.PortMapping{
		portMapping:=strings.Split(pm,":")
		if len(portMapping)!=2{
			log.Fatal("port mapping format error,",pm)
			continue
		}
		// 将宿主机的端口请求转发到容器的地址和端口行
		iptablesCmd:=fmt.Sprintf("-t nat -A PREROUTING -p tcp -m tcp --dport %s -j DNAT --to-destination %s:%s",
			portMapping[0],ep.IpAddress.String(),portMapping[1])

		// 执行iptables命令,添加端口映射转发规则
		cmd:=exec.Command("iptables",strings.Split(iptablesCmd," ")...)
		output,err:=cmd.Output()
		if err!=nil{
			log.Fatal("configPortMapping() iptables Output,",output)
			continue
		}
	}
	return nil
}
