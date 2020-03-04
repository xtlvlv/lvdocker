package command

import (
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
		cli.StringFlag{
			Name:  "m",
			Usage: "内存限制",
		},
	},
	Action: func(ctx *cli.Context) error {
		tty := ctx.Bool("it")
		memory := ctx.String("m")
		command := ctx.Args().Get(0)
		Run(command, tty, memory)
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
