package main

import (
	//	"bytes"
	"flag"
	"fmt"
	"github.com/coreos/go-etcd/etcd"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/http"
	"os"
	//	"strings"
	"strconv"
	"time"
)

var (
	should_alived_host []string
	client             *etcd.Client
	conf               *Config
	ts                 *TSDB
	system_metric_map  map[string]float64
)

func IsFile(file string) bool {
	f, e := os.Stat(file)
	if e != nil {
		return false
	}
	return !f.IsDir()
}

func GetConfig(config_file string) (*Config, error) {
	buf, err := ioutil.ReadFile(config_file)
	if err != nil {
		return nil, err
	}
	conf := Config{}
	err = yaml.Unmarshal(buf, &conf)
	return &conf, err
}

func GetValue(dps map[string]float64) float64 {
	for _, value := range dps {
		return value
	}
	return -1
}

func CheckSysCPU(limit float64) DetectIPMap {
	ret := map[string]float64{}
	for i := 0; i < len(should_alived_host); i++ {
		ip := should_alived_host[i]
		res, _ := ts.Get("system.cpu.sys", "avg", "1m-ago", ip, 0)
		val1 := GetValue(res[0].Dps)
		res, _ = ts.Get("system.cpu.user", "avg", "1m-ago", ip, 0)
		val2 := GetValue(res[0].Dps)
		res, _ = ts.Get("system.cpu.wait", "avg", "1m-ago", ip, 0)
		val3 := GetValue(res[0].Dps)
		val := val1 + val2 + val3
		if val > limit {
			ret[ip] = val
		}
	}
	return ret
}

func CheckSysLoad(limit float64) DetectIPMap {
	ret := map[string]float64{}
	for i := 0; i < len(should_alived_host); i++ {
		ip := should_alived_host[i]
		res, _ := ts.Get("system.load.1m", "avg", "1m-ago", ip, 0)
		load1m := GetValue(res[0].Dps)
		res, _ = ts.Get("system.cpu.cores", "avg", "1m-ago", ip, 0)
		corenum := GetValue(res[0].Dps)
		if load1m > limit*corenum {
			ret[ip] = load1m
		}
	}
	return ret
}

func CheckMetric(metricinfo string, limit float64) DetectIPMap {
	ret := map[string]float64{}
	for i := 0; i < len(should_alived_host); i++ {
		ip := should_alived_host[i]
		res, _ := ts.Get(metricinfo, "avg", "1m-ago", ip, 0)
		val := GetValue(res[0].Dps)
		if val > limit {
			ret[ip] = val
		}
	}
	return ret
}

func MetricInfoInit() {
	system_metric_map = make(map[string]float64)
	system_metric_map["system.mem.percent"] = conf.LimitInfo.Sysmempercent
	system_metric_map["system.disk.system.percent"] = conf.LimitInfo.Diskpercent
	system_metric_map["system.swap.percent"] = conf.LimitInfo.Swap
	system_metric_map["system.cpu.percent"] = conf.LimitInfo.Syscpu
	system_metric_map["sys.load.1m"] = conf.LimitInfo.Sysload

}

type DetectIPMap map[string]float64

func Metric_Check() map[string]DetectIPMap {
	ret := map[string]DetectIPMap{}
	for key, value := range system_metric_map {
		if key == "system.cpu.percent" {
			ret[key] = CheckSysCPU(value)
		} else if key == "sys.load.1m" {
			ret[key] = CheckSysLoad(value)
		} else {
			ret[key] = CheckMetric(key, value)
		}
	}
	return ret
}

func Check() {

	for {
		message := ""
		losted_host := Host_alive_check()
		if len(losted_host) > 0 {
			message += "Host Lost list as below:\n"
			for i := 0; i < len(losted_host); i++ {
				message += losted_host[i] + "\n"
			}
		}

		metricIP := Metric_Check()

		for key, value := range metricIP {
			if len(value) > 0 {
				message += "\n\nFind below host out of limit on metric: " + key + ".\n"
				for key1, value1 := range value {
					message += key1 + "\t" + strconv.FormatFloat(value1, 'f', -1, 64) + "\n"
				}
				message += "\n"
			}
		}

		if message != "" {
			message += "\nsend time: " + time.Now().Format("2006-01-02 15:04:05")
			fmt.Printf("Send message:\n%s", message)
			err := SendMail(conf.Smtp.Host, conf.Smtp.Port, conf.Smtp.Username, conf.Smtp.Password, []string{conf.Smtp.Mailto}, "Error", message)
			if err != nil {
				fmt.Printf("Send mail error: %v\n", err)
			}

		}
		time.Sleep(time.Duration(conf.Internal) * time.Second)
	}
}

func main() {
	config_path := flag.String("c", "", "config file path")
	*config_path = "./config.yml"
	if !IsFile(*config_path) {
		fmt.Printf("Error: No config file find in path: %s. \n", *config_path)
		return
	}
	config_file := *config_path
	tempconf, err := GetConfig(config_file)
	if err != nil {
		fmt.Printf("get config error: %v\n", err)
		return
	}
	conf = tempconf

	should_alived_host = conf.InitHost

	machines := []string{}
	for i := 0; i < len(conf.Etcd_host); i++ {
		machines = append(machines, "http://"+conf.Etcd_host[i])
	}
	client = etcd.NewClient(machines)
	ts = &TSDB{
		host: conf.Tsdb_host,
		port: conf.Tsdb_port,
	}
	MetricInfoInit()

	go Check()
	http.HandleFunc("/host/remove", RemoveIP)
	http.HandleFunc("/host/list", LiveHosts)
	http.HandleFunc("/metrics", MetricAll)
	http.HandleFunc("/metric/system/list", SysMetricList)
	http.HandleFunc("/metric/system/add", SysMetricAdd)
	http.HandleFunc("/metric/system/remove", SysMetricRemove)

	listento := conf.Server.Host + ":" + strconv.Itoa(conf.Server.Port)
	err = http.ListenAndServe(listento, nil)
	if err != nil {
		fmt.Println(err.Error())
	}

}
