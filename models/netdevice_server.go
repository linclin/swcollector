package models

import (
	"github.com/astaxie/beego/orm"
)

type NetdeviceServer struct {
	Id          int    `orm:"colum(id);auto"`
	NetdeviceId int    `orm:"colum(netdevice_id)"`
	ServerId    int    `orm:"colum(server_id)"`
	PortName    string `orm:"colum(port_name);size(255);null"`
}

func (t *NetdeviceServer) TableName() string {
	return "netdevice_server"
}

func init() {
	orm.RegisterModel(new(NetdeviceServer))
}
