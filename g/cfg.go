package g

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/toolkits/file"
)

type DebugmetricConfig struct {
	Endpoints []string `json:"endpoints`
	Metrics   []string `json:"metrics`
	Tags      string   `json:"tags"`
}

type SwitchConfig struct {
	Enabled  bool     `json:"enabled"`
	IpRange  []string `json:"ipRange"`
	Interval int      `json:"interval"`

	PingTimeout int `json:"pingTimeout"`
	PingRetry   int `json:"pingRetry"`

	Community   string `json:"community"`
	SnmpTimeout int    `json:"snmpTimeout"`
	SnmpRetry   int    `json:"snmpRetry"`

	User     string `json:"user"`
	Password string `json:"password"`
	Port     string `json:"port"`

	IgnoreIface        []string `json:"ignoreIface"`
	IgnorePkt          bool     `json:"ignorePkt"`
	IgnoreOperStatus   bool     `json:"ignoreOperStatus"`
	IgnoreBroadcastPkt bool     `json:"ignoreBroadcastPkt"`
	IgnoreMulticastPkt bool     `json:"ignoreMulticastPkt"`
	DisplayByBit       bool     `json:"displayByBit"`
	LimitConcur        int      `json:"limitConcur"`
	FastPingMode       bool     `json:"fastPingMode"`
}

type CMDBConfig struct {
	Host     string `json:"host"`
	User     string `json:"user"`
	Password string `json:"password"`
	Port     string `json:"port"`
	DB       string `json:"dbname"`
}

type UworkConfig struct {
	Host     string `json:"host"`
	User     string `json:"user"`
	Password string `json:"password"`
	Port     string `json:"port"`
	DB       string `json:"dbname"`
}

type FalconConfig struct {
	Host     string `json:"host"`
	User     string `json:"user"`
	Password string `json:"password"`
	Port     string `json:"port"`
	DB       string `json:"dbname"`
}


type HeartbeatConfig struct {
	Enabled  bool   `json:"enabled"`
	Addr     string `json:"addr"`
	Interval int    `json:"interval"`
	Timeout  int    `json:"timeout"`
}


type TransferConfig struct {
	Enabled  bool   `json:"enabled"`
	Addr    string `json:"addr"`
	Interval int    `json:"interval"`
	Timeout  int    `json:"timeout"`
}

type HttpConfig struct {
	Enabled  bool     `json:"enabled"`
	Listen   string   `json:"listen"`
	TrustIps []string `json:trustIps`
}

type SwitchHostsConfig struct {
	Enabled bool   `json:enabled`
	Hosts   string `json:hosts`
}

type CustomMetricsConfig struct {
	Enabled  bool   `json:enbaled`
	Template string `json:template`
}

type GlobalConfig struct {
	Debug     bool             `json:"debug"`
	IP        string           `json:"ip"`
	Hostname  string           `json:"hostname"`
	User      string           `json:"user"`
	Password  string           `json:"password"`
	Community string           `json:"community"`
	Switch    *SwitchConfig    `json:"switch"`
	Cmdb            string      `json:"cmdb"`
	Heartbeat *HeartbeatConfig `json:"heartbeat"`
	Transfer  *TransferConfig  `json:"transfer"`
	Http      *HttpConfig      `json:"http"`
	CMDB      *CMDBConfig      `json:"cmdb"`
	Uwork     *UworkConfig     `json:"uwork"`
	Falcon    *FalconConfig    `json:"falcon"`
	DefaultTags   map[string]string `json:"default_tags"`
	BackupDir string           `json:"backupdir"`
}


var (
	ConfigFile string
	config     *GlobalConfig
	reloadType bool
	lock       = new(sync.RWMutex)
	rlock      = new(sync.RWMutex)
)

func SetReloadType(t bool) {
	rlock.RLock()
	defer rlock.RUnlock()
	reloadType = t
	return
}

func ReloadType() bool {
	rlock.RLock()
	defer rlock.RUnlock()
	return reloadType
}

func Config() *GlobalConfig {
	lock.RLock()
	defer lock.RUnlock()
	return config
}

func ParseConfig(cfg string) {
	if cfg == "" {
		log.Fatalln("use -c to specify configuration file")
	}

	if !file.IsExist(cfg) {
		log.Fatalln("config file:", cfg, "is not existent. maybe you need `mv cfg.example.json cfg.json`")
	}

	ConfigFile = cfg

	configContent, err := file.ToTrimString(cfg)
	if err != nil {
		log.Fatalln("read config file:", cfg, "fail:", err)
	}

	var c GlobalConfig
	err = json.Unmarshal([]byte(configContent), &c)
	if err != nil {
		log.Fatalln("parse config file:", cfg, "fail:", err)
	}

	lock.Lock()
	defer lock.Unlock()

	config = &c
	SetReloadType(false)
	log.Println("read config file:", cfg, "successfully")

}
