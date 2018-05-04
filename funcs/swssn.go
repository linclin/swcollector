package funcs

import (
	"fmt"
	"log"
	"strings"

	"github.com/astaxie/beego/orm"
	"github.com/gaochao1/sw"
	"github.com/gaochao1/swcollector/g"
	_ "github.com/go-sql-driver/mysql"
)

type SwSSN struct {
	Ip           string
	SSN          string
	Manufacturer string
	Model        string
}

func UpdateSwHdInfo() {
	log.Println("Updating SSN")

	o := orm.NewOrm()
	o.Using("default")

	swHdinfo := SwSSNMetrics()
	log.Println(swHdinfo)

	for _, item := range swHdinfo {
		if item.SSN != "" {
			go updateSwHdInfo(item, o)
		}
	}
}

func updateSwHdInfo(swSSN SwSSN, o orm.Ormer) {
	var sw []orm.Params

	num, err := o.Raw(fmt.Sprintf("SELECT Id,Sn,Brand,Model FROM netdevice WHERE ManagementIp = '%s'", swSSN.Ip)).Values(&sw)
	if err != nil {
		log.Println("Cannot find netdevice ", swSSN.Ip, " from mysql")
		return
	}

	if num > 1 {
		log.Println("Update infomation from ", swSSN.Ip, " failed,can not find the only ManagementIp.")
		return
	}

	if num == 1 {
		ID := GetInt(sw[0]["Id"])
		SN := GetString(sw[0]["Sn"])
		Brand := GetString(sw[0]["Brand"])
		Model := GetString(sw[0]["Model"])

		if SN != swSSN.SSN && swSSN.SSN != "" {
			_, err := o.Raw(fmt.Sprintf("UPDATE netdevice SET Sn='%s' WHERE Id='%d'", swSSN.SSN, ID)).Exec()
			if err != nil {
				log.Println("Update Sn failed", swSSN.Ip, err.Error())
			}
		}

		if Brand != swSSN.Manufacturer && swSSN.Manufacturer != "" {
			_, err := o.Raw(fmt.Sprintf("UPDATE netdevice SET Brand='%s' WHERE Id='%d'", swSSN.Manufacturer, ID)).Exec()
			if err != nil {
				log.Println("Update Brand failed", swSSN.Ip, err.Error())
			}
		}

		if Model != swSSN.Model && swSSN.Model != "" {
			_, err := o.Raw(fmt.Sprintf("UPDATE netdevice SET Model='%s' WHERE Id='%d'", swSSN.Model, ID)).Exec()
			if err != nil {
				log.Println("Update Model failed", swSSN.Ip, err.Error())
			}
		}
		return
	}

	if num == 0 {
		_, err := o.Raw(fmt.Sprintf("INSERT INTO netdevice (Sn,Brand,Model,ManagementIp) VALUES ('%s','%s','%s','%s')", swSSN.SSN, swSSN.Manufacturer, swSSN.Model, swSSN.Ip)).Exec()
		if err != nil {
			log.Println("Insert netdevice failed", swSSN.Ip, err.Error())
		}
		return
	}

	return
}

func SwSSNMetrics() (L []SwSSN) {
	chs := make([]chan SwSSN, len(AliveIp))
	for i, ip := range AliveIp {
		if ip != "" {
			chs[i] = make(chan SwSSN)
			go SSNReport(ip, chs[i])
		}
	}

	for _, ch := range chs {
		swSSN := <-ch
		L = append(L, swSSN)
	}

	return L
}

func SSNReport(ip string, ch chan SwSSN) {
	var swSSN SwSSN

	snInfo, Manufacturer := GetSSN(ip)
	swSSN.Ip = ip
	swSSN.SSN = snInfo["SSN"]
	swSSN.Model = snInfo["Model"]
	swSSN.Manufacturer = Manufacturer

	ch <- swSSN

	return

}

func GetSSN(ip string) (hdinfo map[string]string, Manufacturer string) {
	user := g.Config().Switch.User
	password := g.Config().Switch.Password
	ip_port := fmt.Sprintf("%s:%s", ip, g.Config().Switch.Port)

	sysDescr, _ := sw.SysDescr(ip, g.Config().Switch.Community, 15)
	sysDescrLower := strings.ToLower(sysDescr)

	if strings.Contains(sysDescrLower, "cisco") {
		info := SSHCommand(user, password, ip_port, "show inventory")
		hdinfo := CiscoInfo(info)
		return hdinfo, "cisco"
	}
	if strings.Contains(sysDescrLower, "dell") {
		info := SSHCommand(user, password, ip_port, "show version")
		hdinfo := DellInfo(info)
		return hdinfo, "dell"
	}
	if strings.Contains(sysDescrLower, "h3c") {
		info := SSHCommand(user, password, ip_port, "display device manuinfo")
		hdinfo := H3CInfo(info)
		return hdinfo, "h3c"
	}
	if strings.Contains(sysDescrLower, "hillstone") {
		info := SSHCommand(user, password, ip_port, "show version")
		hdinfo := HillstoneInfo(info)
		return hdinfo, "hillstone"
	}
	HDInfo := map[string]string{
		"SSN":   "",
		"Model": "",
	}
	return HDInfo, ""
}

func CiscoInfo(info string) (hdinfo map[string]string) {
	HDInfo := map[string]string{
		"SSN":   "",
		"Model": "",
	}

	if info == "" {
		return HDInfo
	}

	data := strings.Split(info, "\n")[1]
	HDInfo["Model"] = strings.TrimSpace(strings.Split(strings.Split(data, ",")[0], ":")[1])
	HDInfo["SSN"] = strings.TrimSpace(strings.Split(strings.Split(data, ",")[2], ":")[1])

	return HDInfo
}

func DellInfo(info string) (hdinfo map[string]string) {
	HDInfo := map[string]string{
		"SSN":   "",
		"Model": "",
	}

	if info == "" {
		return HDInfo
	}

	data := strings.Split(info, "\r")
	for _, line := range data {
		if strings.Contains(line, "System Model ID") {
			res := strings.Split(line, ".")
			HDInfo["Model"] = res[len(res)-1]
		}
		if strings.Contains(line, "Serial Number") {
			res := strings.Split(line, ".")
			HDInfo["SSN"] = res[len(res)-1]
		}
	}

	return HDInfo
}

func H3CInfo(info string) (hdinfo map[string]string) {
	HDInfo := map[string]string{
		"SSN":   "",
		"Model": "",
	}

	if info == "" {
		return HDInfo
	}

	data := strings.Split(info, "\r")
	HDInfo["Model"] = strings.TrimSpace(strings.Split(data[18], ":")[1])
	HDInfo["SSN"] = strings.TrimSpace(strings.Split(data[20], ":")[1])

	return HDInfo
}

func HillstoneInfo(info string) (hdinfo map[string]string) {
	HDInfo := map[string]string{
		"SSN":   "",
		"Model": "",
	}

	if info == "" {
		return HDInfo
	}

	data := strings.Split(info, "\r")
	HDInfo["SSN"] = strings.TrimSpace(strings.Split(data[3], " ")[4])
	HDInfo["Model"] = strings.TrimSpace(strings.Split(data[3], ":")[2])

	return HDInfo
}
