package models

import (
	"github.com/astaxie/beego/orm"
)

type NetdevicePort struct {
	Id          int    `orm:"colum(id);auto"`
	NetdeviceId int    `orm:"colum(netdevice_id)"`
	PortName    string `orm:"colum(port_name);size(255);null"`
	BindIp      string `orm:"colum(bind_ip);size(255);null"`
}

func (t *NetdevicePort) TableName() string {
	return "netdevice_port"
}

func init() {
	orm.RegisterModel(new(NetdevicePort))
}
