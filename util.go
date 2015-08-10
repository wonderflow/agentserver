// SaveJson project main.go
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

var (
	Logger *log.Logger
)

func SaveJson(w http.ResponseWriter, req *http.Request) {
	var data map[string]string
	buff := new(bytes.Buffer)
	buff.ReadFrom(req.Body)
	err := json.Unmarshal(buff.Bytes(), &data)
	if err != nil {
		Logger.Println(err.Error())
	}
	Jdata, err := json.Marshal(data)
	Logger.Println(string(Jdata))
	if err != nil {
		Logger.Println(err.Error())
	}
	err = ioutil.WriteFile("config.json", Jdata, 0644)
	if err != nil {
		Logger.Println(err.Error())
	}
}
func ReadJson(w http.ResponseWriter, req *http.Request) {
	var data map[string]string
	result, err := ioutil.ReadFile("config.json")
	if err != nil {
		Logger.Println(err.Error())
	}
	err = json.Unmarshal(result, &data)
	if err != nil {
		Logger.Println(err.Error())
	}
	Logger.Println(data)
}

//func main() {
//	logfile, err := os.OpenFile("./agentserver.log", os.O_CREATE|os.O_APPEND|os.O_RDWR, os.ModePerm)
//	if err != nil {
//		fmt.Printf("%s\r\n", err.Error())
//		os.Exit(-1)
//	}
//	defer logfile.Close()
//	Logger = log.New(logfile, "", log.Ldate|log.Ltime|log.Llongfile)

//	http.HandleFunc("/save", SaveJson)
//	http.HandleFunc("/read", ReadJson)
//	err = http.ListenAndServe("0.0.0.0:50000", nil)
//	if err != nil {
//		fmt.Println(err.Error())
//	}
//}
