package controllers

import (
	"github.com/EPICPaaS/appmsgsrv/appweb/models"
	"github.com/astaxie/beego"
)

/*user控制器*/
type UserController struct {
	beego.Controller
}

func (this *UserController) Get() {

	this.TplNames = "login.html"
	this.Data["info"] = "denglu jiemian"
	this.Data["Tenants"] = models.GetAll()
}

func (this *UserController) Post() {
	ret := Result{
		Success: true,
	}
	uname := this.Input().Get("userName")
	tenantId := this.Input().Get("tenantId")
	pwd := this.Input().Get("pasword")
	user := models.GetUserByNameTenantId(uname, tenantId)

	if user != nil && user.Password == pwd {
		ret.Data = user
	} else {
		ret.Success = false
		ret.Msg = "用户名或密码错误"
	}
	this.Data["json"] = ret
	this.ServeJson()
	this.StopRun()

}
