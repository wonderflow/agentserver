package main

type Threshold struct {
	Sysload        float64 `yaml:"sysload"`
	Syscpu         float64 `yaml:"syscpu"`
	Sysmempercent  float64 `yaml:"sysmempercent"`
	Swap           float64 `yaml:"swap"`
	Diskpercent    float64 `yaml:"diskpercent"`
	Router_rps_max float64 `yaml:"Router_rps_max"`
}

type SMTPinfo struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Mailto   string `yaml:"mailto"`
}

type Server struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type Config struct {
	Etcd_host    []string  `yaml:"etcd_host"`
	Etcd_dir     string    `yaml:"etcd_dir"`
	Etcdjobs_dir string    `yaml:"etcdjobs_dir"`
	Etcd_rm_dir  string    `yaml:"etcd_rm_dir"`
	Tsdb_host    string    `yaml:"tsdb_host"`
	Tsdb_port    string    `yaml:"tsdb_port"`
	LimitInfo    Threshold `yaml:"threshold"`
	Internal     int       `yaml:"internal"`
	Load_percent int       `yaml:"load_percent"`
	Smtp         SMTPinfo  `yaml:"smtp"`
	Server       Server    `yaml:"server“`
	InitHost     []string  `yaml:"inithost"`
}
