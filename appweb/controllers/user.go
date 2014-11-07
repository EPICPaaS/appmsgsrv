package controllers

import (
	//"fmt"
	"github.com/EPICPaaS/appmsgsrv/appweb/models"
	"github.com/astaxie/beego"
)

/*user控制器*/
type UserController struct {
	beego.Controller
}

func (this *UserController) Get() {

	this.TplNames = "login.html"
	this.Data["Tenants"] = models.GetAll()
}

func (this *UserController) Post() {

	//返回结果
	ret := Result{
		Success: true,
	}
	uname := this.Input().Get("userName")
	tenantId := this.Input().Get("tenantId")
	pwd := this.Input().Get("password")
	user := models.GetUserByNameTenantId(uname, tenantId)

	if user != nil && user.Password == pwd {
		ret.Data = user
		this.SetSession("user", user)

	} else {
		ret.Success = false
		ret.Msg = "用户名或密码错误"
	}

	this.Data["json"] = ret
	this.ServeJson()
	this.StopRun()
}

func (this *UserController) Logout() {
	//清除缓存
	this.DelSession("user")
	this.Redirect("/", 302)
}
