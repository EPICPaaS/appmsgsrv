package controllers

import (
	"github.com/EPICPaaS/appmsgsrv/appweb/models"
	"github.com/astaxie/beego"
)

type IndexController struct {
	beego.Controller
}

func (this *IndexController) Get() {
	this.TplNames = "index.html"
	user := this.GetSession("user").(*models.User)
	orgs := models.GetRootOrg(user.TenantId)
	this.Data["orgs"] = orgs
}

//获取下一单位
func (this *IndexController) GetChildOrgs() {

	parentId := this.Input().Get("parentId")
	orgs := models.GetChildOrgs(parentId)
	this.Data["json"] = orgs
	this.ServeJson()
	this.StopRun()
}

//获取单位用户
func (this *IndexController) GetOrgUserById() {
	orgId := this.Input().Get("orgId")
	users := models.GetOrgUsersById(orgId)
	this.Data["json"] = users
	this.ServeJson()
	this.StopRun()
}

//获取用户的群
func (this *IndexController) GetMyQun() {

	user := this.GetSession("user").(*models.User)
	this.Data["json"] = models.GetMyQun(user.Id)
	this.ServeJson()
	this.StopRun()

}
