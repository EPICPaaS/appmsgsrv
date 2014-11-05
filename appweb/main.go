package main

import (
	_ "github.com/EPICPaaS/appmsgsrv/appweb/routers"
	"github.com/astaxie/beego"
)

func main() {
	beego.Run()
}

