package models

import (
	"bytes"
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

type Org struct {
	Id        string `orm:"column(id);pk"`
	Name      string
	ShortName string
	ParentId  string
	Location  string
	TenantId  string
	Sort      int
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

//获取一级单位
func GetRootOrg(tenantId string) *[]Org {

	if len(tenantId) == 0 {
		return nil
	}
	o := orm.NewOrm()
	qs := o.QueryTable("org")
	orgs := &[]Org{}
	_, err := qs.Filter("tenant_id", tenantId).Filter("parent_id", "").All(orgs)
	if err != nil {
		beego.Error(err)
		return nil
	}
	return orgs
}

//获取指定单位的直属下级单位
func GetChildOrgs(parentId string) *[]Org {
	if len(parentId) == 0 {
		return nil
	}
	o := orm.NewOrm()
	qs := o.QueryTable("org")
	orgs := &[]Org{}
	_, err := qs.Filter("parent_id", parentId).All(orgs)
	if err != nil {
		beego.Error(err)
		return nil
	}
	return orgs
}

//获取指定单位的所有下级单位
func GetOrgsByParentId(parentId string) *[]Org {
	if len(parentId) == 0 {
		return nil
	}
	o := orm.NewOrm()
	orgs := &[]Org{}
	_, err := o.Raw("select * from (select location from org where id =?) t1 , org t2 where t2.location like CONCAT(t1.location,'%')", parentId).QueryRows(orgs)
	if err != nil {
		beego.Error(err)
		return nil
	}
	return orgs
}

//获取单位下的所有人员
func GetOrgUsersById(orgId string) *[]User {
	if len(orgId) == 0 {
		return nil
	}
	o := orm.NewOrm()
	//查询该单位的所有下级单位
	orgs := GetOrgsByParentId(orgId)
	if orgs != nil {
		users := &[]User{}
		//拼接查询语句
		selectUser := bytes.Buffer{}
		selectUser.WriteString("select from user where ")
		ls := len(*orgs) - 1
		for i, org := range *orgs {
			selectUser.WriteString(" id =" + org.Id)
			if i != ls {
				selectUser.WriteString(" or ")
			}
		}
		_, err := o.Raw("select  * from user where id in (select user_id from org_user where ? )", selectUser.String()).QueryRows(users)

		if err != nil {
			beego.Error(err)
		}
		return users
	}

	return nil
}
