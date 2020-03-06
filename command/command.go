package command

import (
	"fmt"
	"github.com/urfave/cli"
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
	},
	Action: func(ctx *cli.Context) error {
		tty := ctx.Bool("it")
		d:=ctx.Bool("d")
		if d{
			tty=false
		}
		memory := ctx.String("m")
		volume:=ctx.String("v")
		containerName:=ctx.String("name")
		command := ctx.Args().Get(0)

		Run(command, tty, memory,volume,containerName)
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
