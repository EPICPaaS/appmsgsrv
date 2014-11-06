package main

import (
	_ "github.com/EPICPaaS/appmsgsrv/appweb/routers"
	_ "github.com/EPICPaaS/appmsgsrv/appweb/setting"
	"github.com/astaxie/beego"
)

func main() {

	beego.Run()
	beego.SetStaticPath("/static", "static")
}
