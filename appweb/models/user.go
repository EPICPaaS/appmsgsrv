package models

import (
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
	"time"
)

type User struct {
	Id          string `orm:"column(id);pk"`
	Name        string
	Nickname    string
	Avatar      string
	NamePy      string
	NameQuanpin string
	Status      int
	Rand        int
	Password    string
	TenantId    string
	Level       int
	Email       string
	Mobile      string
	Area        string
	Created     time.Time
	Updated     time.Time
}

/*获取*/
func GetUserByNameTenantId(userName, tenantId string) *User {

	var users []User
	o := orm.NewOrm()
	num, err := o.Raw("SELECT * FROM user WHERE name = ? and tenant_id=?", userName, tenantId).QueryRows(&users)
	if err != nil {
		fmt.Println("查询出错")
		beego.Error(err)
		return nil
	}
	if num > 0 {
		return &users[0]
	}
	return nil
}

//验证用户是否登陆
var IsUserLogin = func(ctx *context.Context) {
	user := ctx.Input.Session("user")
	if user == nil && ctx.Request.RequestURI != "/login" {
		ctx.Redirect(302, "/")
	}
}
