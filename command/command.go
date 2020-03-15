package command

import (
	"fmt"
	"github.com/urfave/cli"
	"log"
	"modfinal/network"
)

/*
RunCommand run 命令
*/
var RunCommand = cli.Command{
	Name: "run",
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "it",
			Usage: "指定在当前终端运行",
		},
		cli.BoolFlag{
			Name:        "d",
			Usage:       "后台运行,enable detach",
		},
		cli.StringFlag{
			Name:  "m",
			Usage: "内存限制",
		},
		cli.StringFlag{
			Name:        "v",
			Usage:       "enable volume",
		},
		cli.StringFlag{
			Name:        "name",
			Usage:       "指定容器名字",
		},
		cli.StringFlag{
			Name:        "net",
			Usage:       "container network",
		},
		cli.StringSliceFlag{
			Name:      "p",
			Usage:     "port mapping",
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
		volume:=ctx.String("v")
		containerName:=ctx.String("name")
		command := ctx.Args().Get(0)

		Run(command, tty, memory,volume,containerName,network,portMapping)
		return nil
	},
}

/*
InitCommand init命令,不会自己调用,在Run()里调用
*/
var InitCommand = cli.Command{
	Name: "init",
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
