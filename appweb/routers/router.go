package routers

import (
	"github.com/EPICPaaS/appmsgsrv/appweb/controllers"
	"github.com/astaxie/beego"
)

func init() {
    beego.Router("/", &controllers.MainController{})
}
