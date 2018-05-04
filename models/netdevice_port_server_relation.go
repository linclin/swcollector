package models

import (
	"github.com/astaxie/beego/orm"
)

type NetdevicePortServerRelation struct {
	Id              int    `orm:"colum(id);auto"`
	NetdeviceId     int    `orm:"colum(netdevice_id)"`
	NetdevicePortId int    `orm:"colum(netdevice_port_id)"`
	ServerId        int    `orm:"colum(server_id)"`
	MacAdd          string `orm:"colum(mac_add);size(255);null"`
}

func (t *NetdevicePortServerRelation) TableName() string {
	return "netdevice_port_server_relation"
}

func init() {
	orm.RegisterModel(new(NetdevicePortServerRelation))
}
