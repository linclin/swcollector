package funcs

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"github.com/gaochao1/swcollector/models"
	"github.com/astaxie/beego/orm"
	"github.com/gaochao1/sw"
	"github.com/gaochao1/swcollector/g"
	_ "github.com/go-sql-driver/mysql"
	"unsafe"
	"github.com/open-falcon/common/model"
	"encoding/json"
	"strconv"
	"bytes"
	"io/ioutil"
	"time"
)

type SwPortInfo struct {
	SwIp string
	Ip   string
	Mac  string
	Port string
}
type AdsHostInfoRequest struct {
	HostInfo AdsReportRequest `json:"hostinfo",omitempty`
}
//ADS:auto discovery system
type AdsReportRequest struct {
	AgentVersion string `json:"bk_agent_version,omitempty"` //厂商 Agent版本
	Hostname     string `json:"bk_host_name,omitempty"`     //主机名称
	SN           string `json:"bk_sn,omitempty"`            //主机sn
	Uuid         string `json:"bk_uuid,omitempty"`          //主机sn
	Manufacturer string `json:"bk_manufacturer,omitempty"`  //厂商
	ProductName  string `json:"bk_productName,omitempty"`   //型号
	LanIP        string `json:"bk_host_innerip,omitempty"`  //内网IP地址
	WanIP        string `json:"bk_host_outerip,omitempty"`  //外网IP地址
	//VIP          string  `json:"bk_host_outerip"`//VIP地址
	FwName    string `json:"bk_os_name,omitempty"`    //操作系统名称
	FwVersion string `json:"bk_os_version,omitempty"` //内核版本
	FwType    string `json:"bk_os_type,omitempty"`    //系统类型
	FwBit     string `json:"bk_os_bit,omitempty"`     //操作系统位数

	Mem  uint `json:"bk_mem,omitempty"`  //内存容量
	Disk uint `json:"bk_disk,omitempty"` //磁盘容量

	Cpu       uint  `json:"bk_cpu,omitempty"`        //CPU逻辑核心数
	CpuMhz    uint `json:"bk_cpu_mhz,omitempty"`    //CPU频率
	CpuModule string `json:"bk_cpu_module,omitempty"` //CPU型号

	Mac           string `json:"bk_mac,omitempty"`            //内网MAC地址
	LanMask       string `json:"bk_lan_mask,omitempty"`       //内网掩码
	LanGateway    string `json:"bk_lan_gateway,omitempty"`    //内网网关
	OuterMac      string `json:"bk_outer_mac,omitempty"`      //外网MAC
	OuterMask     string `json:"bk_outer_mask,omitempty"`     //外网掩码
	OuterGateway  string `json:"bk_outer_gateway,omitempty"`  //外网网关
	HostManageIp  string `json:"bk_host_manageip,omitempty"`  //带外IP
	ManageMask    string `json:"bk_manage_mask,omitempty"`    //带外掩码
	ManageGateway string `json:"bk_manage_gateway,omitempty"` //带外网关

	UpdateTime string `json:"bk_agent_update_time,omitempty"` //更新时间
	ImportFrom string `json:"import_from,omitempty"` //录入方式
	HostType string `json:"bk_host_type,omitempty"` //主机类型

}
type AdsReportRes struct {
	Result           bool `json:"result,omitempty"`
	BkErrorCode           int `json:"bk_error_code,omitempty"`
	BkErrorMsg           string `json:"bk_error_msg,omitempty"`
}


var  AllSwPortInfo  []SwPortInfo

//交换机基础信息采集上报
func UpdateSwUpstreamPort() {
	log.Println("Updating SSN")
	o := orm.NewOrm()
	o.Using("default")

	//更新硬件信息
	swHdinfo := SwSSNMetrics()
	log.Println(swHdinfo)
	GetAdsInfo(swHdinfo)

	AllSwPortInfo = make ([]SwPortInfo,0)
	//更新端口信息
	portData := SwPort()
	for _, portInfo := range portData {
		go updateUpstreamPortInfo(portInfo, o)
	}

	log.Println("AllSwPortInfo is " ,AllSwPortInfo)
}

//更新端口信息
func updateUpstreamPortInfo(portInfo PortInfo, o orm.Ormer) {
	log.Println("Starting update portinfo,portInfo Ip is ", portInfo.Ip)
	//使用sshCommond 上机器查端口
	user := g.Config().Switch.User
	password := g.Config().Switch.Password
	ip_port := fmt.Sprintf("%s:%s", portInfo.Ip, g.Config().Switch.Port)
	sysDescr, _ := sw.SysDescr(portInfo.Ip, g.Config().Switch.Community, 15)
	sysDescrLower := strings.ToLower(sysDescr)


	if strings.Contains(sysDescrLower, "cisco") {
		info := SSHCommand(user, password, ip_port, "sho ip arp | include 10.205")
		data := strings.Split(info, "\n")
		for i, in := range data {
			log.Println("detail: index ", i, "data:", in, "port id is ", portInfo.Ip)
			//ip:data2[0]  mac:data2[5]  port_name:strings.Split(portInfos," ")[len(strings.Split(portInfos," "))-1]
			data2 := strings.Split(in, " ")
			if len(data2) > 1 {
				portInfos := SSHCommand(user, password, ip_port, "sho mac address-table |include "+data2[5])
				SwPortInfo := SwPortInfo{}
				SwPortInfo.SwIp = portInfo.Ip
				SwPortInfo.Ip = data2[0]
				SwPortInfo.Mac = data2[5]
				SwPortInfo.Port = strings.Split(portInfos, " ")[len(strings.Split(portInfos, " "))-1]
				AllSwPortInfo = append(AllSwPortInfo, SwPortInfo)
			}
		}
	}

	if strings.Contains(sysDescrLower, "h3c") {
		info := SSHCommand(user, password, ip_port, "dis arp | include 10.205")
		data := strings.Split(info, "\n")
		log.Println("length of data is ", len(data))
		for _, in := range data {
			if strings.Contains(in, "10.205.") {
				//log.Println("detail: index ", i, "data:", in, "port id is ", portInfo.Ip)
				data2 := strings.Split(in, " ")
				//ip : data2[0] mac_add :data2[3]  port_name :data2[12]
				if len(data2) > 1 {
					mac_add := data2[3]
					port_name := data2[12]
					if data[3] == "" {
						mac_add = data2[2]
					}
					if data[12] == "" {
						port_name = data2[11]
					}

					SwPortInfo := SwPortInfo{}
					SwPortInfo.SwIp = portInfo.Ip
					SwPortInfo.Ip = data2[0]
					SwPortInfo.Mac = mac_add
					SwPortInfo.Port = port_name
					AllSwPortInfo = append(AllSwPortInfo, SwPortInfo)
				}
			}
		}
	}
	return
}

//更新交换机信息进数据库
func updateSwUpstreamPort(swSSN SwSSN, o orm.Ormer) {
	var sw []orm.Params

	num, err := o.Raw(fmt.Sprintf("SELECT Id,Sn,Brand,Model FROM netdevice WHERE Sn = '%s'", swSSN.SSN)).Values(&sw)
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
		//SN := GetString(sw[0]["Sn"])
		//Brand := GetString(sw[0]["Brand"])
		//Model := GetString(sw[0]["Model"])
		//更新网段
		ManIP := GetString(sw[0]["ManagementIp"])

		if ManIP != swSSN.Ip && swSSN.Ip != "" {
			_, err := o.Raw(fmt.Sprintf("UPDATE netdevice SET ManagementIp='%s' WHERE Id='%d'", swSSN.Ip, ID)).Exec()
			if err != nil {
				log.Println("Update ManagementIp failed", swSSN.Ip, err.Error())
			}
		}

		//if SN != swSSN.SSN && swSSN.SSN != "" {
		//	_, err := o.Raw(fmt.Sprintf("UPDATE netdevice SET Sn='%s' WHERE Id='%d'", swSSN.SSN, ID)).Exec()
		//	if err != nil {
		//		log.Println("Update Sn failed", swSSN.Ip, err.Error())
		//	}
		//}

		//if Brand != swSSN.Manufacturer && swSSN.Manufacturer != "" {
		//	_, err := o.Raw(fmt.Sprintf("UPDATE netdevice SET Brand='%s' WHERE Id='%d'", swSSN.Manufacturer, ID)).Exec()
		//	if err != nil {
		//		log.Println("Update Brand failed", swSSN.Ip, err.Error())
		//	}
		//}

		//if Model != swSSN.Model && swSSN.Model != "" {
		//	_, err := o.Raw(fmt.Sprintf("UPDATE netdevice SET Model='%s' WHERE Id='%d'", swSSN.Model, ID)).Exec()
		//	if err != nil {
		//		log.Println("Update Model failed", swSSN.Ip, err.Error())
		//	}
		//}
		return
	}

	if num == 0 {
		log.Println("inster swSSN.Ip", swSSN.Ip, swSSN.SSN)
		_, err := o.Raw(fmt.Sprintf("INSERT INTO netdevice (Sn,Brand,Model,ManagementIp) VALUES ('%s','%s','%s','%s')", swSSN.SSN, swSSN.Manufacturer, swSSN.Model, swSSN.Ip)).Exec()
		if err != nil {
			log.Println("Insert netdevice failed", swSSN.Ip, err.Error())
		}
		return
	}

	return
}

//更新端口与服务对应信息
func updatePortServerRelation(o orm.Ormer) {
	var allPort []orm.Params
	_, err := o.Raw("SELECT id,netdevice_id,mac_add FROM netdevice_port WHERE mac_add != ''  ").Values(&allPort)
	if err != nil {
		log.Println("Search failure", err.Error())
		return
	}
	for _, ss := range allPort {
		macAdd := ""
		if strings.Contains(GetString(ss["mac_add"]), "-") {
			//turns A-B-C to a:b:c:d:e:f
			macAddArray := strings.Split(GetString(ss["mac_add"]), "-")
			for i := 0; i < len(macAddArray); i++ {
				macAdd += string(macAddArray[i][0]) + string(macAddArray[i][1]) + ":" + string(macAddArray[i][2]) + string(macAddArray[i][3]) + ":"
			}
			macAdd = strings.TrimRight(macAdd, ":")
		}
		if strings.Contains(GetString(ss["mac_add"]), ".") {
			//turns A.B.C to a:b:c:d:e:f
			macAddArray := strings.Split(GetString(ss["mac_add"]), ".")
			for i := 0; i < len(macAddArray); i++ {
				macAdd += string(macAddArray[i][0]) + string(macAddArray[i][1]) + ":" + string(macAddArray[i][2]) + string(macAddArray[i][3]) + ":"
			}
			macAdd = strings.TrimRight(macAdd, ":")
		}

		if macAdd != "" {
			var result []orm.Params
			num, err1 := o.Raw("SELECT SvrId FROM server_hard_nic WHERE NicMacAddress = ?  ", macAdd).Values(&result)
			if err1 != nil || num == 0 {
				continue
			} else if num == 1 {
				//更新或者插入
				line := models.NetdevicePortServerRelation{NetdeviceId: GetInt(ss["id"]), NetdevicePortId: GetInt(ss["netdevice_id"]), ServerId: GetInt(result[0]["SvrId"]), MacAdd: macAdd}
				if create, _, line_err := o.ReadOrCreate(&line, "mac_add"); line_err == nil {
					if create {
						log.Println("Update successed:", macAdd)
					} else {
						log.Println("Alreadly existed:", macAdd)
					}
				} else {
					log.Println("Update failed:", macAdd, line_err.Error())
				}
			}
		}
	}
}


func IsIntranet(ipStr string) bool {
	if strings.HasPrefix(ipStr, "10.") || strings.HasPrefix(ipStr, "192.168.") {
		return true
	}

	if strings.HasPrefix(ipStr, "172.") {
		// 172.16.0.0-172.31.255.255
		arr := strings.Split(ipStr, ".")
		if len(arr) != 4 {
			return false
		}

		second, err := strconv.ParseInt(arr[1], 10, 64)
		if err != nil {
			return false
		}

		if second >= 16 && second <= 31 {
			return true
		}
	}

	return false
}

func GetAdsInfo(swHdinfo []SwSSN) string {
	for _,ss :=range swHdinfo{
		req := AdsReportRequest{}
		req.AgentVersion=g.VERSION
		req.ImportFrom ="2" //上报方式
		req.UpdateTime = time.Now().Format("2006-01-02 15:04:05")
		req.SN = ss.SSN
		req.Manufacturer= ss.Manufacturer
		req.ProductName=ss.Model
		req.HostType ="5"
		req.HostManageIp = ss.Ip
		metricValue,_:=reportAdsToCmdb(req)
		mvs := []*model.MetricValue{}
		mvs=append(mvs,&metricValue)
		g.SendToTransfer(mvs)
		return ""
	}

	return ""
}

func reportAdsToCmdb(req AdsReportRequest)  (model.MetricValue,error){
	metricValue := model.MetricValue{
		Endpoint:req.ProductName+"/"+req.HostManageIp,
		Metric:"ads.server",
		Value:0,
		Step:3600,
		Type:"GAUGE",
		Timestamp:time.Now().Unix(),
	}


	if g.Config().Cmdb != "" {
		adsHostInfoRequest:=AdsHostInfoRequest{
			HostInfo:req,
		}
		bytesData, err := json.Marshal(adsHostInfoRequest)
		if err != nil {
			fmt.Println(err.Error() )
			return metricValue,err
		}
		fmt.Println("上报数据")
		fmt.Println(string(bytesData))
		reader := bytes.NewReader(bytesData)
		url :=g.Config().Cmdb
		request, err := http.NewRequest("POST", url, reader)
		if err != nil {
			fmt.Println(err.Error())
			return metricValue,err
		}
		request.Header.Set("Content-Type", "application/json;charset=UTF-8")
		request.Header.Set("BK_USER", "admin")
		request.Header.Set("HTTP_BLUEKING_SUPPLIER_ID", "0")
		client := http.Client{}
		resp, err := client.Do(request)
		if err != nil {
			fmt.Println(err.Error())
			return metricValue,err
		}
		respBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err.Error())
			return metricValue,err
		}
		//byte数组直接转成string，优化内存
		str := (*string)(unsafe.Pointer(&respBytes))
		fmt.Println("上报结果")
		fmt.Println(*str)
		res := AdsReportRes{}
		err =json.Unmarshal(respBytes,&res)
		fmt.Println(res)
		if err == nil {
			if res.Result {
				metricValue.Value=1
			}
		}
	}
	return metricValue,nil
}
