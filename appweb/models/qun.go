package models

import (
	//"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
	"time"
)

type Qun struct {
	Id          string `orm:"column(id);pk"`
	CreatorId   string
	Name        string
	Description string
	MaxMember   string
	Avatar      string
	Created     time.Time
	Updated     time.Time
}

type QunUser struct {
	Id      string `orm:"column(id);pk"`
	QunId   string
	UserId  string
	Sort    int
	role    int
	Created time.Time
	Updated time.Time
}

//获取用户所参与的群
func GetMyQun(userId string) *[]Qun {
	if len(userId) == 0 {
		return nil
	}
	o := orm.NewOrm()
	quns := &[]Qun{}
	_, err := o.Raw("select * from qun where id in (SELECT qun_id from qun_user where user_id = '?')", userId).QueryRows(quns)
	if err != nil {
		beego.Error(err)
		return nil
	}
	return quns
}
