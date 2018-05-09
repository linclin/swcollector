package g

import (
	"log"
	"os"
	"strings"
	"sync"

	"github.com/toolkits/slice"

	"time"

	"github.com/open-falcon/common/model"
	"github.com/toolkits/net"
	"bytes"
)

var Root string

func InitRootDir() {
	var err error
	Root, err = os.Getwd()
	if err != nil {
		log.Fatalln("getwd fail:", err)
	}
}

var LocalIps []string

func InitLocalIps() {
	var err error
	LocalIps, err = net.IntranetIP()
	if err != nil {
		log.Fatalln("get intranet ip fail:", err)
	}
}

var (
	TransferClient *SingleConnRpcClient
)

func InitRpcClients() {
	if Config().Transfer.Enabled {
		TransferClient = &SingleConnRpcClient{
			RpcServer: Config().Transfer.Addr,
			Timeout:   time.Duration(Config().Transfer.Timeout) * time.Millisecond,
		}
	}
}

func SendToTransfer(metrics []*model.MetricValue) {
	if len(metrics) == 0 {
		return
	}

	dt := Config().DefaultTags
	if len(dt) > 0 {
		var buf bytes.Buffer
		default_tags_list := []string{}
		for k, v := range dt {
			buf.Reset()
			buf.WriteString(k)
			buf.WriteString("=")
			buf.WriteString(v)
			default_tags_list = append(default_tags_list, buf.String())
		}
		default_tags := strings.Join(default_tags_list, ",")

		for i, x := range metrics {
			buf.Reset()
			if x.Tags == "" {
				metrics[i].Tags = default_tags
			} else {
				buf.WriteString(metrics[i].Tags)
				buf.WriteString(",")
				buf.WriteString(default_tags)
				metrics[i].Tags = buf.String()
			}
		}
	}

	debug := Config().Debug

	if debug {
		log.Printf("=> <Total=%d> %v\n", len(metrics), metrics[0])
	}

	var resp model.TransferResponse
	SendMetrics(metrics, &resp)

	if debug {
		log.Println("<=", &resp)
	}
}

func array_include(array_a []string, array_b []string) bool { //b include a
	for _, v := range array_a {
		if in_array(v, array_b) {
			continue
		} else {
			return false
		}
	}
	return true
}

func in_array(a string, array []string) bool {
	for _, v := range array {
		if a == v {
			return true
		}
	}
	return false
}

var (
	ips     []string
	ipsLock = new(sync.Mutex)
)

func TrustableIps() []string {
	ipsLock.Lock()
	defer ipsLock.Unlock()
	ips := Config().Http.TrustIps
	return ips
}

func IsTrustable(remoteAddr string) bool {
	ip := remoteAddr
	idx := strings.LastIndex(remoteAddr, ":")
	if idx > 0 {
		ip = remoteAddr[0:idx]
	}

	if ip == "127.0.0.1" {
		return true
	}

	return slice.ContainsString(TrustableIps(), ip)
}
