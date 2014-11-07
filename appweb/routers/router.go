package routers

import (
	"github.com/EPICPaaS/appmsgsrv/appweb/controllers"
	"github.com/astaxie/beego"
)

func init() {
	beego.Router("/", &controllers.UserController{})
	beego.Router("/login", &controllers.UserController{})
	beego.Router("/index", &controllers.IndexController{})
}
