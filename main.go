package main

import (
	//	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/coreos/go-etcd/etcd"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	//	"strings"
	//	"net/url"
	"time"
)

var (
	should_alived_host []string
	should_alived_jobs []string
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

func GetValueFromRes(res Response) float64 {
	if len(res) > 0 {
		for _, value := range res[0].Dps {
			return value
		}
	}
	return -1
}

func CheckSysCPU(limit float64) DetectIPMap {
	ret := map[string]float64{}
	for i := 0; i < len(should_alived_host); i++ {
		ip := should_alived_host[i]
		res, _ := ts.Get("system.cpu.sys", "avg", "1m-ago", ip, 0)
		val1 := GetValueFromRes(res)
		if val1 < 0 {
			continue
		}
		res, _ = ts.Get("system.cpu.user", "avg", "1m-ago", ip, 0)
		val2 := GetValueFromRes(res)
		if val2 < 0 {
			continue
		}
		res, _ = ts.Get("system.cpu.wait", "avg", "1m-ago", ip, 0)
		val3 := GetValueFromRes(res)
		if val3 < 0 {
			continue
		}
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
		load1m := GetValueFromRes(res)
		res, _ = ts.Get("system.cpu.cores", "avg", "1m-ago", ip, 0)
		corenum := GetValueFromRes(res)
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
		res, err := ts.Get(metricinfo, "avg", "1m-ago", ip, 0)
		if err != nil {
			fmt.Println("Get metric info error: ", err)
			continue
		}

		//fmt.Println(res)
		val := GetValueFromRes(res)
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

func SysMetric_Check() map[string]DetectIPMap {
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

func isAlive(val string) bool {
	for _, host := range should_alived_host {
		if val == host {
			return true
		}
	}
	return false
}

type Job struct {
	Jobname string
	IP      string
	Index   int
}

func GetJobList() (jobs []Job) {
	resp, err := client.Get(conf.Etcdjobs_dir, true, true)
	if err != nil {
		fmt.Println("Get etcd job list error: ", err)
	}

	for _, val := range resp.Node.Nodes {
		subres, err := client.Get(val.Key, true, true)
		if err != nil {
			fmt.Println("Get etcd job list error: ", err)
		}

		for _, subval := range subres.Node.Nodes {
			tmpJob := Job{}
			json.Unmarshal([]byte(subval.Value), &tmpJob)
			jobs = append(jobs, tmpJob)
		}
	}
	return
}

func CheckProcessMonitor(jobs []Job) (unmonitoredjobs []Job) {
	for _, val := range jobs {
		res, _ := ts.Get("process.process.monitor", "avg", "1m-ago", val.Jobname, val.Index)
		monitor := GetValueFromRes(res)
		if monitor != 1 {
			unmonitoredjobs = append(unmonitoredjobs, val)
		}
	}
	return
}

func CheckProcessStatus(jobs []Job) (unstatedjobs []Job) {
	for _, val := range jobs {
		res, _ := ts.Get("process.process.status", "avg", "1m-ago", val.Jobname, val.Index)

		status := GetValueFromRes(res)
		if status != 0 {
			unstatedjobs = append(unstatedjobs, val)
		}
	}
	return
}

func CheckProcessMem(jobs []Job) (outmemjobs []Job) {
	for _, val := range jobs {
		res, _ := ts.Get("process.cpu.percent", "avg", "1m-ago", val.Jobname, val.Index)
		mem := GetValueFromRes(res)
		if mem > conf.LimitInfo.Sysmempercent { //todo: 暂时用系统的limit？
			outmemjobs = append(outmemjobs, val)
		}
	}
	return
}

func CheckJobList(jobs []Job) (losted_jobs []string) {
	for _, val := range should_alived_jobs {
		find := false
		for _, job := range jobs {
			if job.Jobname == val {
				find = true
				break
			}
		}
		if !find {
			losted_jobs = append(losted_jobs, val)
		}
	}
	return
}

func ProcessMetric_Check() (process_msg string) {
	process_msg = ""
	jobs := GetJobList()
	losted_jobs := CheckJobList(jobs)

	if len(losted_jobs) > 0 {
		process_msg += "\nJobs losted: \n"
	}
	for _, val := range losted_jobs {
		process_msg += val + "\n"
	}

	unmonitoredjobs := CheckProcessMonitor(jobs)
	if len(unmonitoredjobs) > 0 {
		process_msg += "\nJobs unmonitored: \n"
		for _, val := range unmonitoredjobs {
			process_msg += val.Jobname + "\t" + strconv.Itoa(val.Index)
		}
	}

	unstatedjobs := CheckProcessStatus(jobs)

	if len(unstatedjobs) > 0 {
		process_msg += "\nJobs status wrong: \n"
		for _, val := range unstatedjobs {
			process_msg += val.Jobname + "\t" + strconv.Itoa(val.Index)
		}
	}

	outmemjobs := CheckProcessMem(jobs)

	if len(outmemjobs) > 0 {
		process_msg += "\nJobs out of memory: \n"
		for _, val := range outmemjobs {
			process_msg += val.Jobname + "\t" + strconv.Itoa(val.Index)
		}

	}

	//TODO
	// checkMonitorStatus => process.process.monitor  == 1 //done
	// checkProcessStatus => process.process.status  == 0 // done to be test
	// checkProcessCpu => process.cpu.percenttotal 这个不能超过配置
	// process.mem.percent
	return process_msg
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

		metricIP := SysMetric_Check()

		ProcessMetric_Check()

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
	should_alived_jobs = conf.InitJobs

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
	//	AppCheck()
	go Check()
	http.HandleFunc("/remove", Remove)
	http.HandleFunc("/list", Alived)
	http.HandleFunc("/metrics", MetricAll)
	http.HandleFunc("/metric/system/list", SysMetricList)
	http.HandleFunc("/metric/system/add", SysMetricAdd)
	http.HandleFunc("/metric/system/remove", SysMetricRemove)
	http.HandleFunc("/app/list", AppCheck)

	listento := conf.Server.Host + ":" + strconv.Itoa(conf.Server.Port)
	err = http.ListenAndServe(listento, nil)
	if err != nil {
		fmt.Println(err.Error())
	}

}
