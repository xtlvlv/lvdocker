package main

import (
"log"
"os"
"modfinal/command"
"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "lvdocker"
	app.Usage = "类docker容器引擎"
	app.Commands = []cli.Command{
		command.RunCommand,
		command.InitCommand,
		command.ListCommand,
		command.LogsCommand,
		command.StopCommand,
		command.RemoveCommand,
		command.CommitCommand,
		command.NetworkCommand,
		command.WebCommand,
	}
	log.Println("lvdocker 开始运行")
	if err := app.Run(os.Args); err != nil {
		log.Fatal("main.go error", err)
	}
}
