package network

import (
	"fmt"
	"github.com/vishvananda/netlink"
	"log"
	"net"
	"os/exec"
	"strings"
	"time"
)

type BridgeNetworkDriver struct {

}


/*
一些公共方法
*/

/*
initBridge相当于
ip link add bridgeName type bridge
ip addr add ip dev bridgeName
ip link set bridgeName up
初始化bridge的四个流程
1. 创建Bridge虚拟设备
2. 设置Bridge设备地址和路由
3. 启动Bridge设备
4. 设置iptables SNAT规则,保证挂载到这个Bridge上的容器的Veth能够访问外部网络
 */
func (d *BridgeNetworkDriver) initBridge(n *Network) error {
	//1. 创建Bridge虚拟设备
	bridgeName:=n.Name
	createBridgeInterface(bridgeName)

	//2. 设置Bridge设备地址和路由
	gatewayIp:=*n.IpRange
	gatewayIp.IP=n.IpRange.IP
	setInterfaceIp(bridgeName,gatewayIp.String())

	//3. 启动Bridge设备
	setInterfaceUp(bridgeName)

	//4. 设置iptables SNAT规则,保证挂载到这个Bridge上的容器的Veth能够访问外部网络
	setupIpTables(bridgeName,n.IpRange)

	return nil
}

func (d *BridgeNetworkDriver) deleteBridge(n *Network) error {
	bridgeName:=n.Name

	l,err:=netlink.LinkByName(bridgeName)
	if err!=nil{
		log.Fatal("deleteBridge() error ",err)
	}
	if err:=netlink.LinkDel(l);err!=nil{
		log.Fatal("deleteBridge() error2 ",err)
	}
	return nil
}

/*
创建bridge虚拟设备
 */
func createBridgeInterface(bridgeName string) error {

	// 先检查是否存在同名bridge设备
	_,err:=net.InterfaceByName(bridgeName)
	// 若已经存在或报错
	if err==nil||!strings.Contains(err.Error(),"no such network interface"){
		log.Println("bridge 已存在,bridgeName:",bridgeName)
		return err
	}

	// 初始化一个netlink的Link基础对象,Link的名字即Bridge虚拟设备的名字
	la:=netlink.NewLinkAttrs()
	la.Name=bridgeName

	// 使用刚才创建的Link的属性创建netlink的Bridge对象
	br:=&netlink.Bridge{
		LinkAttrs:         la,
		MulticastSnooping: nil,
		HelloTime:         nil,
		VlanFiltering:     nil,
	}

	// 调用netlink的Linkadd方法,创建Bridge虚拟网络设备,相当于 ip link add xxx
	if err:=netlink.LinkAdd(br);err!=nil{
		log.Fatal("创建bridge失败,createBridgeInterface(),",err)
	}

	return nil
}

/*
启动bridge设备
设置网络接口为UP状态
 */
func setInterfaceUp(interfaceName string) error {

	iface,err:=netlink.LinkByName(interfaceName)
	if err!=nil{
		log.Fatal("bridgeDriver.go setInterfaceUp,",err)
	}
	// 等价于 ip link set xxx up 命令
	if err:=netlink.LinkSetUp(iface);err!=nil{
		log.Fatal("bridgeDriver.go setInterfaceUp2,",err)
	}
	return nil
}

/*
设置一个网络接口的IP地址,例如setInterfaceIp("testBridge","192.168.0.1/24")
 */
func setInterfaceIp(name string,rowIp string) error {
	retries:=2
	var iface netlink.Link
	var err error
	for i:=0;i<retries;i++{
		iface,err=netlink.LinkByName(name)
		if err==nil{
			break
		}
		log.Println("netlink link error,retrying...")
		time.Sleep(2*time.Second)
	}
	if err!=nil{
		log.Fatal("bridgeDriver.go setInterfaceIp() link error,",err)
	}
	ipNet,err:=netlink.ParseIPNet(rowIp)
	if err!=nil{
		log.Fatal("setInterface() error2,",err)
	}
	addr:=&netlink.Addr{
		IPNet:       ipNet,
		Label:       "",
		Flags:       0,
		Scope:       0,
		Peer:        nil,
	}
	return netlink.AddrAdd(iface,addr)
}

/*
设置iptables Linux Bridge SNAT规则
设置iptables对应bridge的MASQUERADE规则
 */
func setupIpTables(bridgeName string,subnet *net.IPNet) error {
	// 由于go语言没有直接操控iptables操作的库,所以需要通过命令方式来配置
	// iptables -t nat -A POSTROUTING -s <bridgeName> ! -o <bridgeName> -j MASQUERADE
	iptablesCmd:=fmt.Sprintf("-t nat -A POSTROUTING -s %s ! -o %s -j MASQUERADE",subnet.String(),bridgeName)
	cmd:=exec.Command("iptables",strings.Split(iptablesCmd," ")...)
	output,err:=cmd.Output()
	if err!=nil{
		log.Fatal("setupIpTables() error,",output,err)
	}
	return nil
}

func (d *BridgeNetworkDriver) Name() string {
	return "bridge"
}

/*
返回一个Network,该Network会有一个网桥,名字为传入的name,网段为传入的网段subnet,并且该网络已经up并且已经设置好iptables的MASQUERADE规则
 */
func (d *BridgeNetworkDriver) Create(subnet string,name string) (*Network,error) {
	ip,ipRange,_:=net.ParseCIDR(subnet)	// 这个函数功能是把网段的字符串转换成net.IPnet的对象
	ipRange.IP=ip
	n:=&Network{
		Name:       name,
		IpRange:    ipRange,
		DriverName: d.Name(),
	}
	err:=d.initBridge(n)
	if err!=nil{
		log.Fatal("bridgeDriver.go,",err)
	}
	return n,err
}

/*
删除该网络的网桥设备
相当于 ip link delete bridgeName type bridge
 */
func (d *BridgeNetworkDriver) Delete(network Network) error {

	bridgeName:=network.Name
	br,err:=netlink.LinkByName(bridgeName)
	if err!=nil{
		log.Fatal("Delete() 查找网桥失败,",err)
	}
	if err=netlink.LinkDel(br);err!=nil{
		log.Fatal("Delete() 删除网桥失败,",err)
	}
	return nil
}

/*
关联设备
connect方法生成一个设备Veth并且将其中一端attach搭配该网络的网桥上
 */
func (d *BridgeNetworkDriver) Connect(network *Network,endpoint *EndPoint) error {

	bridgeName:=network.Name
	br,err:=netlink.LinkByName(bridgeName)
	if err!=nil{
		log.Fatal("Connect() 查找网桥失败,",err)
	}

	// 创建Veth接口的配置
	la:=netlink.NewLinkAttrs()

	// 由于Linux接口名的限制,名字取endpoint ID的前5位
	la.Name=endpoint.Id[:5]

	// 通过设置Veth接口的master属性,设置这个Veth的一端挂载到网络对应的Linux Bridge上
	la.MasterIndex=br.Attrs().Index

	// 创建Veth对象,通过PeerName配置Veth另外一端的接口名
	endpoint.Device=netlink.Veth{
		LinkAttrs:        la,
		PeerName:         "cif-"+endpoint.Id[:5],
		PeerHardwareAddr: nil,
	}

	// 通过netlink的LinkAdd方法创建出这个Veth接口
	// 因为上面指定了link的MasterIndex是网络对应的Linux Bridge
	// 所以Veth的一端就已经挂载到了网络对应的Linux Bridge上
	if err=netlink.LinkAdd(&endpoint.Device);err!=nil{
		log.Fatal("Connect() linkadd error,",err)
	}
	// 设置Veth启动, ip link set xxx up
	if err=netlink.LinkSetUp(&endpoint.Device);err!=nil{
		log.Fatal("Connect() LinkSetUp error,",err)
	}
	return nil
}

func (d *BridgeNetworkDriver) Disconnect(network Network,endpoint *EndPoint) error{
	return nil
}