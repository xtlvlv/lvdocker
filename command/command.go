package command

import (
	"fmt"
	"github.com/urfave/cli"
	"log"
	"modfinal/cgroups"
	"modfinal/cgroups/subsystems"
	"modfinal/network"
)

/*
RunCommand run 命令
*/
var RunCommand = cli.Command{
	Name: "run",
	Usage:"run命令",
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "it",
			Usage: "指定在当前终端运行",
		},
		cli.BoolFlag{
			Name:        "d",
			Usage:       "后台运行,enable detach,指定后台才有日志",
		},
		cli.StringFlag{
			Name:  "m",
			Usage: "内存限制,格式为16m",
		},
		cli.StringFlag{
			Name:  "c",
			Usage: "cpuset(核心数)限制,格式为0-7",
		},
		cli.StringFlag{
			Name:        "v",
			Usage:       "enable volume,指定宿主机的数据卷与容器目录",
		},
		cli.StringFlag{
			Name:        "name",
			Usage:       "指定容器名字",
		},
		cli.StringFlag{
			Name:        "image",
			Usage:       "指定容器镜像,如果不指定就用默认的busybox",
		},
		cli.StringFlag{
			Name:        "net",
			Usage:       "container network",
		},
		cli.StringSliceFlag{
			Name:      "p",
			Usage:     "port mapping,指定端口映射",
		},
	},
	Action: func(ctx *cli.Context) error {
		tty := ctx.Bool("it")
		network := ctx.String("net")
		portMapping := ctx.StringSlice("p")
		d:=ctx.Bool("d")
		if d{
			tty=false
		}
		memory := ctx.String("m")
		cpu:=ctx.String("c")
		res:=subsystems.ResourceConfig{
			MemoryLimit:memory,
			CpuLimit:cpu,
		}
		cg:=cgroups.CgroupManager{
			Resource:      &res,
			SubsystemsIns: make([]subsystems.Subsystem,0),
		}
		if memory!=""{
			cg.SubsystemsIns=append(cg.SubsystemsIns,&subsystems.MemorySubsystem{})
		}
		if cpu!=""{
			cg.SubsystemsIns=append(cg.SubsystemsIns,&subsystems.CpuSubSystem{})
		}
		volume:=ctx.String("v")
		containerName:=ctx.String("name")
		imageName:=ctx.String("image")
		command := ctx.Args().Get(0)

		Run(command, tty, cg,volume,containerName,imageName,network,portMapping)
		return nil
	},
}

/*
InitCommand init命令,不会自己调用,在Run()里调用
*/
var InitCommand = cli.Command{
	Name: "init",
	Usage:"初始化,主动调用没有意义",
	Action: func(ctx *cli.Context) error {
		//command := ctx.Args().Get(0)
		//Init(command)
		Init()
		return nil
	},
}

/*
list命令,列出容器信息
 */
var ListCommand=cli.Command{
	Name:                   "ps",
	Usage:"展示当前宿主机运行的容器",
	Action: func(ctx *cli.Context) error{
		List()
		return nil
	},
}

/*
logs命令,查看容器日志
 */
var LogsCommand  =  cli.Command{
	Name:                   "logs",
	Usage:"加上容器名字,查看该容器的日志,适用于后台运行的容器",
	Action: func(ctx *cli.Context) error{
		if len(ctx.Args())<1{
			fmt.Println("参数过少,请加上容器名字")
			return nil
		}
		containerName:=ctx.Args().Get(0)
		Logs(containerName)
		return nil
	},

}

/*
stop命令,停止容器
 */
var StopCommand=cli.Command{
	Name:                   "stop",
	Usage:"只是停止容器,没有删除",
	Action: func(ctx *cli.Context) error{
		if len(ctx.Args())<1{
			fmt.Println("缺少容器名字")
			return nil
		}
		containerName:=ctx.Args().Get(0)
		Stop(containerName,false)	//如果是前台程序,退出后要用户自己保证不调用stop,实际上调用也没事,只是报错退出
		return nil
	},
}

/*
rm命令,删除容器
 */
var RemoveCommand  = cli.Command{
	Name:                   "rm",
	Usage:"删除容器,从宿主机删除对应的文件夹",
	Action: func(ctx *cli.Context) error{
		if len(ctx.Args())<1{
			fmt.Println("缺少容器名字")
			return nil
		}
		containerName:=ctx.Args().Get(0)
		Remove(containerName)
		return nil
	},
}

/*
commit命令,保存镜像
 */
var CommitCommand=cli.Command{
	Name:                   "commit",
	Usage:"保存镜像,把某个容器的mnt目录打包放到镜像仓库",
	Action: func(ctx *cli.Context) error{
		if len(ctx.Args())<2{
			fmt.Println("缺少容器名字")
			return nil
		}
		containerName:=ctx.Args().Get(0)
		imageName:=ctx.Args().Get(1)
		Commit(containerName,imageName)
		return nil
	},
}

/*
network命令,操作网络
 */
var NetworkCommand=cli.Command{
	Name:                   "network",
	Usage:"通过子命令,创建删除展示网络",
	Subcommands:[]cli.Command{
		{
			Name:"create",
			Flags:[]cli.Flag{
				cli.StringFlag{
					Name:        "driver",
					Usage:"network driver",
				},
				cli.StringFlag{
					Name:        "subnet",
					Usage:"subnet driver",
				},
			},
			Action: func(context *cli.Context) error {
				if len(context.Args())<1{
					log.Fatal("缺少参数:network name")
				}
				network.Init()
				err:=network.CreateNetwork(context.String("driver"),context.String("subnet"),context.Args()[0])
				if err!=nil{
					log.Fatal("command.go network create 失败,",err)
				}
				return nil
			},
		},
		{
			Name:"list",
			Action: func(context *cli.Context) error {

				network.Init()
				network.ListNetwork()
				return nil
			},
		},
		{
			Name:"remove",
			Action: func(context *cli.Context) error {
				if len(context.Args())<1{
					log.Fatal("缺少参数:network name")
				}
				network.Init()
				err:=network.DeleteNetwork(context.Args()[0])
				if err!=nil{
					log.Fatal("command.go network remove 失败,",err)
				}
				return nil
			},
		},

	},
}

/*
web命令,打开浏览器
*/
var WebCommand  = cli.Command{
	Name:                   "web",
	Usage:"可视化查看宿主机容器运行状态",
	Action: func(ctx *cli.Context) error{
		web()
		return nil
	},
}
