package funcs

import (
	"errors"
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/astaxie/beego/orm"
	"github.com/gaochao1/swcollector/g"
	"github.com/gaochao1/swcollector/models"
	_ "github.com/go-sql-driver/mysql"
)

type PortInfo struct {
	Ip   string
	port map[string]string
}

func UpdatePortInfo() {
	log.Println("Updating CMDB")

	o := orm.NewOrm()
	o.Using("default")

	portData := SwPort()

	log.Println(portData)

	for _, portInfo := range portData {
		go updatePortInfo(portInfo, o)
	}
}

func updatePortInfo(portInfo PortInfo, o orm.Ormer) {
	var server []orm.Params

	//log.Println("Starting update portinfo")

	num, err := o.Raw(fmt.Sprintf("SELECT Id FROM netdevice WHERE ManagementIp = '%s'", portInfo.Ip)).Values(&server)

	if err != nil {
		log.Println("Search ", portInfo.Ip, " from mysql failed")
		return
	}

	if num != 1 {
		log.Println("Update infomation from ", portInfo.Ip, " failed,can not find the only ManagementIp.")
		return
	}

	netdeviceId := GetInt(server[0]["Id"])
	//log.Println(netdeviceId)

	for ip, port := range portInfo.port {
		line := models.NetdevicePort{NetdeviceId: netdeviceId, PortName: port, BindIp: ip}
		if create, _, line_err := o.ReadOrCreate(&line, "NetdeviceId", "PortName", "BindIp"); line_err == nil {
			if create {
				log.Println("Update successed:", netdeviceId, port, ip)
			} else {
				log.Println("Alreadly existed:", netdeviceId, port, ip)
			}
		} else {
			log.Println("Update failed:", netdeviceId, port, ip, line_err.Error())
		}
	}
}

func SwPort() (L []PortInfo) {
	//log.Println("Start search the AliveIp", AliveIp)

	chs := make([]chan PortInfo, len(AliveIp))
	for i, ip := range AliveIp {
		if ip != "" {
			chs[i] = make(chan PortInfo)
			go swport(ip, chs[i])
		}
	}

	for _, ch := range chs {
		portInfo := <-ch
		if portInfo.Ip != "" && portInfo.port != nil {
			L = append(L, portInfo)
		} else {
			continue
		}
	}

	return L
}

func swport(ip string, ch chan PortInfo) {
	var portInfo PortInfo

	ifIndex, indeErr := getIfIndex(ip)
	if indeErr != nil {
		log.Println("getIfIndex Error:", indeErr.Error())
		portInfo.Ip = ""
		portInfo.port = nil
		ch <- portInfo
		return
	}

	ifDescr, DescrErr := getIfDescr(ip)
	if DescrErr != nil {
		log.Println("getIfDescr Error:", DescrErr.Error())
		portInfo.Ip = ""
		portInfo.port = nil
		ch <- portInfo
		return
	}

	ifBind := make(map[string]string)

	for key, value := range ifIndex {
		tmp, ok := ifDescr[key]
		if ok {
			ifBind[value] = tmp
		}
	}

	log.Println("Get ifBind from ", ip, ifBind)
	portInfo.Ip = ip
	portInfo.port = ifBind
	ch <- portInfo

	return
}

func getIfIndex(ip string) (ifIndex map[string]string, err error) {
	IfIndex := make(map[string]string)

	cmd := exec.Command("/bin/bash", "-c", fmt.Sprintf("snmpwalk -v 2c -c %s %s ipAdEntIfIndex", g.Config().Community, ip))
	//log.Println(fmt.Sprintf("snmpwalk -v 2c -c %s %s ipAdEntIfIndex", g.Config().Community, ip))
	res, err := cmd.Output()

	if err != nil {
		return nil, errors.New(err.Error())
	}
	data := ""
	for _, i := range string(res) {
		data = data + string(i)
	}
	info := strings.Split(data, "\n")

	for _, line := range info {
		if line != "" {
			interger := strings.TrimSpace(strings.Split(strings.Split(line, "=")[1], ":")[1])
			ip := strings.TrimSpace(strings.Split(strings.Split(line, "=")[0], "ipAdEntIfIndex.")[1])
			IfIndex[interger] = ip
		}
	}
	return IfIndex, nil

}

func getIfDescr(ip string) (ifDescr map[string]string, err error) {
	IfDescr := make(map[string]string)

	cmd := exec.Command("/bin/bash", "-c", fmt.Sprintf("snmpwalk -v 2c -c %s %s ifDescr", g.Config().Community, ip))
	//log.Println(fmt.Sprintf("snmpwalk -v 2c -c %s %s ifDescr", g.Config().Community, ip))
	res, err := cmd.Output()

	if err != nil {
		return nil, errors.New(err.Error())
	}
	data := ""
	for _, i := range string(res) {
		data = data + string(i)
	}
	info := strings.Split(data, "\n")

	for _, line := range info {
		if line != "" {
			interger := strings.TrimSpace(strings.Split(strings.Split(line, "=")[0], ".")[1])
			port := strings.TrimSpace(strings.Split(strings.Split(line, "::")[1], "STRING:")[1])
			IfDescr[interger] = port
		}
	}
	return IfDescr, nil

}
