package controllers

import (
	"github.com/astaxie/beego"
)

/*user控制器*/
type UserController struct {
	beego.Controller
}

func (this *UserController) Get() {
	this.Data["info"] = "denglu jiemian"
	this.Data["Email"] = "astaxie@gmail.com"
	this.TplNames = "index.tpl"
}
