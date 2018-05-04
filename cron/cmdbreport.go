package cron

import (
	"github.com/astaxie/beego/toolbox"
	"github.com/gaochao1/swcollector/funcs"
)

func CMDBReport() {
	//更新交换机上联端口信息
	SwUpdateSwUpstreamPort := toolbox.NewTask("UpdateSwUpstreamPort", "0 0 */3 * * *", func() error {
		go funcs.UpdateSwUpstreamPort()
		return nil
	})
	toolbox.AddTask("UpdateSwUpstreamPort", SwUpdateSwUpstreamPort)
	toolbox.StartTask()
}
