package funcs

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"github.com/gaochao1/sw"
	"github.com/gaochao1/swcollector/g"
	_ "github.com/go-sql-driver/mysql"
	"unsafe"
	"github.com/open-falcon/common/model"
	"encoding/json"
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
	SN           string `json:"bk_sn,omitempty"`            //主机sn
	Manufacturer string `json:"bk_manufacturer,omitempty"`  //厂商
	ProductName  string `json:"bk_productName,omitempty"`   //型号

	HostManageIp  string `json:"bk_host_manageip,omitempty"`  //带外IP
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

	//更新硬件信息
	swHdinfo := SwSSNMetrics()
	log.Println(swHdinfo)
	GetAdsInfo(swHdinfo)

	AllSwPortInfo = make ([]SwPortInfo,0)
	//更新端口信息
	portData := SwPort()
	for _, portInfo := range portData {
		go updateUpstreamPortInfo(portInfo)
	}
}

//更新端口信息
func updateUpstreamPortInfo(portInfo PortInfo) {
	log.Println("Starting update portinfo,portInfo Ip is ", portInfo.Ip)
	//使用sshCommond 上机器查端口
	user := g.Config().Switch.User
	password := g.Config().Switch.Password
	ip_port := fmt.Sprintf("%s:%s", portInfo.Ip, g.Config().Switch.Port)
	sysDescr, _ := sw.SysDescr(portInfo.Ip, g.Config().Switch.Community, 15,3000)
	sysDescrLower := strings.ToLower(sysDescr)


	if strings.Contains(sysDescrLower, "cisco") {
		info := SSHCommand(user, password, ip_port, "sho ip arp | include "+g.Config().Switch.SearchNet)
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
		info := SSHCommand(user, password, ip_port, "dis arp | include "+g.Config().Switch.SearchNet)
		data := strings.Split(info, "\n")
		log.Println("length of data is ", len(data))
		for _, in := range data {
			if strings.Contains(in, g.Config().Switch.SearchNet+".") {
				//log.Println("detail: index ", i, "data:", in, "port id is ", portInfo.Ip)
				data2 := strings.Split(in, " ")
				//ip : data2[0] mac_add :data2[3]  port_name :data2[12]
				if len(data2) > 1 {
					log.Println(len(data2))
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


func GetAdsInfo(swHdinfo []SwSSN) string {
	for _,ss :=range swHdinfo{
		if ss.SSN!=""{
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
		}
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
