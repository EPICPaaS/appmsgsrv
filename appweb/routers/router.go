package routers

import (
	"github.com/EPICPaaS/appmsgsrv/appweb/controllers"
	"github.com/astaxie/beego"
)

func init() {
	beego.Router("/", &controllers.UserController{})
	beego.Router("/login", &controllers.UserController{})
	beego.Router("/logout", &controllers.UserController{}, "get:Logout")
	beego.Router("/index", &controllers.IndexController{})
	beego.Router("/GetOrgUser", &controllers.IndexController{}, "post:GetOrgUserById")
	beego.Router("/GetMyQun", &controllers.IndexController{}, "post:GetMyQun")
	beego.Router("/GetChildOrgs", &controllers.IndexController{}, "post:GetChildOrgs")
}
