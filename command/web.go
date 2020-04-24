package command

import (
	"log"
	"os/exec"
)

func web(){
	url:="http://127.0.0.1:8080/"
	cmd:=exec.Command("xdg-open",url)
	err :=cmd.Start()
	if err!=nil{
		log.Fatal(err)
	}
	log.Println("web run")
	//time.Sleep(time.Second*5)
	err =cmd.Wait()

	if err!=nil{
		log.Fatal("wait ",err)
	}
}