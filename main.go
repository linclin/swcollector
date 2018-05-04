package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/gaochao1/swcollector/cron"
	"github.com/gaochao1/swcollector/funcs"
	"github.com/gaochao1/swcollector/g"
	"github.com/gaochao1/swcollector/http"
	"github.com/astaxie/beego/orm"
)



func init() {
	cfg := flag.String("c", "cfg.json", "configuration file")
	version := flag.Bool("v", false, "show version")
	check := flag.Bool("check", false, "check collector")

	flag.Parse()

	if *version {
		fmt.Println(g.VERSION)
		os.Exit(0)
	}

	g.ParseConfig(*cfg)

	g.InitRootDir()
	g.InitLocalIps()
	g.InitRpcClients()

	if *check {
		funcs.CheckCollector()
		os.Exit(0)
	}

	dbUser := g.Config().CMDB.User
	dbPass := g.Config().CMDB.Password
	dbHost := g.Config().CMDB.Host
	dbPort := g.Config().CMDB.Port
	dbName := g.Config().CMDB.DB
	maxIdleConn := 30
	maxOpenConn := 100
	dbLink := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8", dbUser, dbPass, dbHost, dbPort, dbName) + "&loc=Asia%2FShanghai"
	orm.RegisterDriver("mysql", orm.DRMySQL)
	orm.RegisterDataBase("default", "mysql", dbLink, maxIdleConn, maxOpenConn)
	orm.Debug = false
}

func main() {
	funcs.BuildMappers()
	cron.Collect()
	cron.CMDBReport()

	go http.Start()

	select {}

}

