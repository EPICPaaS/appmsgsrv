package app

import (
	"time"

	"github.com/EPICPaaS/appmsgsrv/db"
	//"github.com/EPICPaaS/go-uuid/uuid"
	"github.com/golang/glog"
)

const (

	// 租户资源插入 SQL.
	InsertResourceSQL = "INSERT INTO `resource` (`id`, `tenant_id`, `name`, `description`, `type`, `content`, `created`, `updated`) VALUES " +
		"(?, ?, ?, ?, ?, ?, ?, ?)"

	// 根据资源id 获取资源
	SelectResourceByIdSQL = "SELECT * FROM `resource` where `id` = ?"

	// 根据租户 id 查询租户的资源.
	SelectResourceByTenantIdSQL = "SELECT * FROM `resource` where `tenant_id`  = ?"

	//根据id修改资源
	UpdateResourceByIdSQL = "UPDATE `resource` SET `tenant_id` = ? ,`name` = ? , `description` = ? , `type` = ? ,`content` = ? ,  `created` = ?  , `updated` = ?   WHERE `id` = ?"

	//根据id删除资源
	DelResourceByIdSQL = "DELETE FROM `resource` WHERE `id` =? "
)

//租户证书资源结构
type Resource struct {
	Id          string    `json:"id"`
	TenantId    string    `json:"tenantId"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Type        string    `json:"type"`
	Content     string    `json:"content"`
	Created     time.Time `json:"created"`
	Updated     time.Time `json:"updated"`
}

// 在数据库中查询资源.
func GetResourceById(resourceId string) (*Resource, error) {
	row := db.MySQL.QueryRow(SelectResourceByIdSQL, resourceId)

	resource := Resource{}
	if err := row.Scan(&resource.Id, &resource.TenantId, &resource.Name, &resource.Description, &resource.Type, &resource.Content, &resource.Created, &resource.Updated); err != nil {
		glog.Error(err)

		return nil, err
	}

	return &resource, nil
}

// 在数据库中查询资源.
func GetResourceByTenantId(tenantId string) ([]*Resource, error) {

	rows, _ := db.MySQL.Query(SelectResourceByTenantIdSQL, tenantId)

	ret := []*Resource{}
	for rows.Next() {
		resource := &Resource{}
		if err := rows.Scan(&resource.Id, &resource.TenantId, &resource.Name, &resource.Description, &resource.Type, &resource.Content, &resource.Created, &resource.Updated); err != nil {
			glog.Error(err)

			return nil, err
		}
		ret = append(ret, resource)
	}

	return ret, nil
}

// 数据库中插入资源
func AddResource(resource *Resource) (*Resource, bool) {
	tx, err := db.MySQL.Begin()

	if err != nil {
		glog.Error(err)
		return nil, false
	}

	// 创建资源记录
	_, err = tx.Exec(InsertResourceSQL, resource.Id, resource.TenantId, resource.Name, resource.Description, resource.Type, resource.Content, resource.Created, resource.Updated)
	if err != nil {
		glog.Error(err)

		if err := tx.Rollback(); err != nil {
			glog.Error(err)
		}
		return nil, false
	}

	if err := tx.Commit(); err != nil {
		glog.Error(err)
		return nil, false
	}
	return resource, true
}

// 数据库中插入资源
func UpdateResource(resource *Resource) (*Resource, bool) {
	tx, err := db.MySQL.Begin()

	if err != nil {
		glog.Error(err)
		return nil, false
	}

	// 创建资源记录
	_, err = tx.Exec(UpdateResourceByIdSQL, resource.TenantId, resource.Name, resource.Description, resource.Type, resource.Content, resource.Created, resource.Updated, resource.Id)
	if err != nil {
		glog.Error(err)

		if err := tx.Rollback(); err != nil {
			glog.Error(err)
		}
		return nil, false
	}

	if err := tx.Commit(); err != nil {
		glog.Error(err)
		return nil, false
	}
	return resource, true
}

//删除资源
func DeleteResourceById(resourceId string) bool {
	tx, err := db.MySQL.Begin()

	if err != nil {
		glog.Error(err)

		return false
	}

	_, err = tx.Exec(DelResourceByIdSQL, resourceId)
	if err != nil {
		glog.Error(err)

		if err := tx.Rollback(); err != nil {
			glog.Error(err)
		}

		return false
	}

	if err := tx.Commit(); err != nil {
		glog.Error(err)

		return false
	}

	return true
}
