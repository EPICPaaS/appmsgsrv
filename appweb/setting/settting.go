package setting

import (
	"github.com/EPICPaaS/appmsgsrv/appweb/models"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
)

func init() {
	// 注册模型
	orm.Debug = true
	orm.RegisterModel(new(models.User), new(models.Tenant), new(models.Org))
	RegisterDB()
	InitSession()
	beego.InsertFilter("/index", beego.BeforeExec, models.IsUserLogin)

}

func RegisterDB() {

	driverName := beego.AppConfig.String("driverName")
	dataSource := beego.AppConfig.String("dataSource")
	maxIdle, _ := beego.AppConfig.Int("maxIdle")
	maxOpen, _ := beego.AppConfig.Int("maxOpen")
	// 注册驱动
	orm.RegisterDriver("mysql", orm.DR_MySQL)
	// 注册默认数据库
	// set default database
	err := orm.RegisterDataBase("default", driverName, dataSource, maxIdle, maxOpen)
	orm.RunCommand()
	//不自动建表
	err = orm.RunSyncdb("default", false, false)
	if err != nil {
		beego.Error(err)
	}
}

func InitSession() {
	beego.SessionOn = true
	beego.SessionProvider = "memory"
	beego.SessionGCMaxLifetime = 600 //60 seconds
	beego.SessionName = "yxUser"
	beego.SessionCookieLifeTime = 600 //60 seconds
	beego.SessionAutoSetCookie = true
	beego.SessionSavePath = "/"
}
