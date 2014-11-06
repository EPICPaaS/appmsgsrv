package models

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
	"time"
)

type Tenant struct {
	Id         string `orm:"column(id);pk"`
	Code       string
	Name       string
	Status     int
	CustomerId string
	Created    time.Time
	Updated    time.Time
}

func GetAll() *[]Tenant {

	o := orm.NewOrm()
	tenants := make([]Tenant, 0)
	qs := o.QueryTable("tenant")
	_, err := qs.All(&tenants)
	if err != nil {
		beego.Error(err)
		return nil
	}
	return &tenants
}
