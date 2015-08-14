package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func Remove(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	req.ParseForm()
	ip := req.Form.Get("ip")
	if ip == "" {
		ip = req.Form.Get("host")
	}

	jobname := req.Form.Get("job")

	if ip == "" && jobname == "" {
		http.Error(w, `{"errorMessage":"`+fmt.Sprintf("No host/ip or job set = %s", ip)+`"}`, 406)
		return
	}
	var find bool
	tt := ""
	if ip != "" {
		find = false
		for i := 0; i < len(should_alived_host); i++ {
			if should_alived_host[i] == ip {
				find = true
				break
			}
		}
		if find == false {
			http.Error(w, `{"errorMessage":" No such IP find."}`, 406)
			return
		}
		response, err := client.AddChild(conf.Etcd_rm_dir, ip, uint64(conf.Internal*3))
		if err != nil {
			http.Error(w, `{"errorMessage":" Server connect etcd error.`+fmt.Sprintf("%v", err)+`"}`, 406)
			return
		}
		tt = "Agent node removed in agentserver, Please shutdown your agent, if not, it will auto added again in " + strconv.Itoa(conf.Internal*2) + "s.\nEtcd response : " + fmt.Sprintf("%v", response.Node)
	}

	if jobname != "" {
		find = false
		for i := 0; i < len(should_alived_jobs); i++ {
			if should_alived_jobs[i] == jobname {
				find = true
				break
			}
		}
		if find == false {
			http.Error(w, `{"errorMessage":" No such IP find."}`, 406)
			return
		}
		response, err := client.AddChild(conf.Etcd_rm_dir, jobname, uint64(conf.Internal*3))
		if err != nil {
			http.Error(w, `{"errorMessage":" Server connect etcd error.`+fmt.Sprintf("%v", err)+`"}`, 406)
			return
		}
		tt = "Job node removed in agentserver, Please shutdown your job, if not, it will auto added again in " + strconv.Itoa(conf.Internal*2) + "s.\nEtcd response : " + fmt.Sprintf("%v", response.Node)
	}

	fmt.Println(tt)
	w.WriteHeader(200)
	w.Write([]byte(tt))
}

func AppCheck(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	v := url.Values{}
	v.Set("grant_type", "password")
	v.Set("username", "admin")
	v.Set("password", "c1oudc0w")
	token := GetAuthToken(v)

	tt := ""

	Apps := ListCFAPP(token)
	for _, val := range Apps.Resources {
		ins := GetAppInstances(val, token)
		tt += val.Entity.Name + "\t" + val.Entity.State + "\t" + strconv.Itoa(ins) + "/" + strconv.Itoa(val.Entity.Instances) + "\n"
	}
	fmt.Println(tt)
	w.WriteHeader(200)
	w.Write([]byte(tt))

}

func Alived(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	req.ParseForm()
	tp := req.Form.Get("type")
	tt := ""
	hoststr := "Standard hosts list as below:\n"
	jobstr := "Standard jobs list as below:"
	for i := 0; i < len(should_alived_host); i++ {
		hoststr += should_alived_host[i] + "\n"
	}
	for i := 0; i < len(should_alived_jobs); i++ {
		jobstr += should_alived_jobs[i] + "\n"
	}

	if tp == "host" || tp == "ip" {
		tt += hoststr
	} else if tp == "job" {
		tt += jobstr
	} else {
		tt += hoststr
		tt += jobstr
	}
	fmt.Println(tt)
	w.WriteHeader(200)
	w.Write([]byte(tt))
}

func SysMetricList(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	res, err := json.Marshal(system_metric_map)
	if err != nil {
		http.Error(w, `{"errorMessage":" List metric error.`+fmt.Sprintf("%v", err)+`"}`, 406)
		return
	}
	fmt.Println(string(res))
	w.Write(res)
}

func SysMetricAdd(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	err := req.ParseForm()
	if err != nil {
		fmt.Println(err)
		http.Error(w, `{"errorMessage":"Request ParseForm error: `+fmt.Sprintf("%v", err)+`"}`, 406)
		return
	}

	metric_name := req.Form.Get("metric")
	metric_limit := req.Form.Get("limit")

	if metric_name == "" || metric_limit == "" {
		http.Error(w, `{"errorMessage":" you haven't set metric or limit ."}`, 406)
		return
	}

	str := strings.Split(metric_name, ".")[0]

	if str == "process" {
		//大伟这里需要加入 tag, 比如 job = dea_next,index=0
	}

	metrics, err := ts.ListMetricsWithQ(500, str)
	if err != nil {
		fmt.Println(err)
		http.Error(w, `{"errorMessage":" Match metric error:`+fmt.Sprintf("%v", err)+`"}`, 406)
		return
	}

	find := false
	for _, value := range metrics {
		if value == metric_name {
			find = true
			break
		}
	}

	if find == false {
		http.Error(w, `{"errorMessage":"metric `+fmt.Sprintf("%v", metric_name)+`didn't find error."}`, 406)
		return
	}

	limit_val, err := strconv.ParseFloat(metric_limit, 64)
	if err != nil || limit_val < 0 {
		http.Error(w, `{"errorMessage":"Parse metric limit error : `+fmt.Sprintf("%v", err)+` or metric limit must greater than 0 : `+fmt.Sprintf("%v", limit_val)+`. "}`, 406)
		return
	}

	tt := "Metric limit not extended yet."

	if str == "system" {
		system_metric_map[metric_name] = limit_val
		tt = "Set metric: " + metric_name + " limit to " + metric_limit + " success!."
		w.Write([]byte(tt))
	} else if str == "process" {
		//大伟，搞起这里
	} else {
		http.Error(w, `{"errorMessage":"Metric limit not extended yet."}`, 406)
	}

}

func SysMetricRemove(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	err := req.ParseForm()
	if err != nil {
		fmt.Println(err)
		http.Error(w, `{"errorMessage":"Request ParseForm error: `+fmt.Sprintf("%v", err)+`"}`, 406)
		return
	}

	metric_name := req.Form.Get("metric")

	find := false

	for key, _ := range system_metric_map {
		if key == metric_name {
			find = true
			break
		}
	}
	if find == false {
		http.Error(w, `{"errorMessage":"Metric didn't find :`+fmt.Sprintf("%v", metric_name)+`."}`, 406)
		return
	}
	delete(system_metric_map, metric_name)
	tt := metric_name + "is deleted."
	fmt.Println(tt)
	w.Write([]byte(tt))
}

func MetricAll(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	err := req.ParseForm()
	if err != nil {
		fmt.Println(err)
		http.Error(w, `{"errorMessage":"Parse request form error.`+fmt.Sprintf("%v", err)+`"}`, 406)
		return
	}
	respnum := req.Form.Get("num")
	q := req.Form.Get("q")
	var num int

	tt := ""

	if respnum != "" {
		num, err = strconv.Atoi(respnum)
		if err != nil {
			fmt.Println(err)
			http.Error(w, `{"errorMessage":"Transfer number to int error.`+fmt.Sprintf("%v", err)+`"}`, 406)
			return
		}
	} else {
		tt += "You can set num = [metrics list number], default show 25 metrics.\n"
		num = 25
	}
	var metrics []string

	if q == "" {
		metrics, err = ts.ListMetrics(num)
	} else {
		metrics, err = ts.ListMetricsWithQ(num, q)
	}
	if err != nil {
		fmt.Println(err)
		http.Error(w, `{"errorMessage":" List metric error.`+fmt.Sprintf("%v", err)+`"}`, 406)
		return
	}
	for i := 0; i < len(metrics); i++ {
		tt += metrics[i] + "\n"
	}
	fmt.Println(tt)
	w.Write([]byte(tt))
}
