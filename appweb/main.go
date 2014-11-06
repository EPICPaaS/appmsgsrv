package main

import (
	_ "github.com/EPICPaaS/appmsgsrv/appweb/routers"
	_ "github.com/EPICPaaS/appmsgsrv/appweb/setting"
	"github.com/astaxie/beego"
	"runtime"
)

func main() {

	runtime.GOMAXPROCS(runtime.NumCPU())
	beego.Run()
	beego.SetStaticPath("/static", "static")
}
