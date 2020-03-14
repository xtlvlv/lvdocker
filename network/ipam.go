package network

import (
	"encoding/json"
	"log"
	"net"
	"os"
	"path"
	"strings"
)

const ipamDefaultAllocatorPath="/home/lvkou/E/Task/毕业设计/root/network/ipam/subnet.json"

// 存放IP地址分配信息
type IPAM struct {
	SubnetAllocatorPath string	// 分配文件存放位置
	Subnets *map[string]string	// 网段和位图算法的数组map,key是网段,value是分配的位图数组
}

/*
初始化一个IPAM对象
 */
var ipAllocator=&IPAM{
	SubnetAllocatorPath: ipamDefaultAllocatorPath,
}

/*
dump()存储网段地址分配信息,持久化Subnets
 */
func (ipam *IPAM) dump() error {

	// 将目录和文件分离开
	ipamConfigFileDir,_:=path.Split(ipamDefaultAllocatorPath)
	// 如果目录不存在就创建
	if _,err:=os.Stat(ipamConfigFileDir);err!=nil{
		if os.IsNotExist(err){
			os.MkdirAll(ipamConfigFileDir,0644)
		}else{
			log.Fatal("ipam dump error,创建文件夹失败,",err)
		}
	}

	// 如果文件不存在就创建
	subnetConfigFile,err:=os.OpenFile(ipamDefaultAllocatorPath,os.O_TRUNC|os.O_CREATE|os.O_WRONLY,0644)
	defer subnetConfigFile.Close()
	if err!=nil{
		log.Fatal("ipam 配置文件打开失败,",err)
	}

	ipamConfigJson,err:=json.Marshal(ipam.Subnets)
	if err!=nil{
		log.Fatal("ipam 写入文件失败,",err)
	}
	_, err =subnetConfigFile.Write(ipamConfigJson)
	if err!=nil{
		log.Fatal("ipam 写入文件失败2,",err)
	}
	return nil
}

/*
load把subnet.json文件中的内容加载到IPAM的Subnets中
 */
func (ipam *IPAM) load() error {

	// 如果文件不存在,就直接返回
	if _,err:=os.Stat(ipam.SubnetAllocatorPath);err!=nil{
		if os.IsNotExist(err){
			return nil
		}else{
			log.Fatal("ipam dump error,创建文件夹失败,",err)
		}
	}

	subnetConfigFile,err:=os.Open(ipam.SubnetAllocatorPath)
	defer subnetConfigFile.Close()
	if err!=nil{
		log.Fatal("ipam 配置文件load失败,",err)
	}

	subnetJson:=make([]byte,2000)
	num, err :=subnetConfigFile.Read(subnetJson)
	if err!=nil{
		log.Fatal("ipam 读出失败,",err)
	}
	log.Println(subnetJson)

	err =json.Unmarshal(subnetJson[:num],ipam.Subnets)
	if err!=nil{
		log.Fatal("ipam json 解析失败,",err)
	}
	return nil
}

/*
地址分配,在网段中分配一个可用的ip地址
1. 先从subnet.json中加载数据到ipam的subnets,如果该文件不存在,subnets是一个空map,里面什么网络信息都没有
2. 根据bitmap分配ip
3. 将已经有数据的subnets持久化到subnet,json中
 */
func (ipam *IPAM) Allocate(subnet *net.IPNet) (ip net.IP,err error) {
	// 存放网段中地址分配信息的数组
	// 无论ipamDefaultAllocatorPath是否存在都先new一个
	ipam.Subnets=&map[string]string{}
	ipam.load()

	// 得到网络号
	_,subnet,_=net.ParseCIDR(subnet.String())

	log.Printf("Allocate subnet:%s, ipam.Subnets:%v\n",subnet,ipam.Subnets)

	// one 表示前缀的个数,size 表示ip地址的个数, ipv4==>size=32
	one,size:=subnet.Mask.Size()
	log.Printf("Allocate one:%d, size:%d\n",one,size)

	// 如果该网络还不在ipam.Subnet中,则初始化一个
	// size-one表示主机号占用位数,2^(size-one)就是主机ip个数
	if _,exist:=(*ipam.Subnets)[subnet.String()];!exist{
		// 1<<uint8(size-one) 等价于 2^(size-one),这么多全是0的字符串,哪个被分配就改成1
		(*ipam.Subnets)[subnet.String()]=strings.Repeat("0",1<<uint8(size-one))
	}

	log.Printf("Allocate one:%s\n",(*ipam.Subnets)[subnet.String()])

	// 遍历网段的位图数组
	for c:=range ((*ipam.Subnets)[subnet.String()]){
		// 如果第c个ip没有被分配 则分配
		if (*ipam.Subnets)[subnet.String()][c]=='0'{
			ipalloc:=[]byte((*ipam.Subnets)[subnet.String()])
			ipalloc[c]='1'
			(*ipam.Subnets)[subnet.String()]=string(ipalloc)

			// 查一下c 用32位如何表示
			// 这里的IP为初始IP,如对于网段192.168.0.0/16,这里的IP就是192.168.0.0
			ip=subnet.IP
			for t:=uint(4);t>0;t-=1{
				[]byte(ip)[4-t]+=uint8(c>>((t-1)*8))
			}
			ip[3]+=1
			break
		}
	}

	ipam.dump()
	return
}

/*
地址释放
 */

func (ipam *IPAM) Release(subnet *net.IPNet,ipaddr *net.IP) error {
	ipam.Subnets=&map[string]string{}

	// 得到网络号
	_,subnet,_=net.ParseCIDR(subnet.String())
	ipam.load()

	c:=0
	releaseIP:=ipaddr.To4()
	releaseIP[3]-=1
	for t:=uint(4);t>0;t-=1{
		c+=int(releaseIP[t-1]-subnet.IP[t-1])<<((4-t)*8)
	}

	ipalloc:=[]byte((*ipam.Subnets)[subnet.String()])
	ipalloc[c]='0'
	(*ipam.Subnets)[subnet.String()]=string(ipalloc)
	ipam.dump()
	return nil
}