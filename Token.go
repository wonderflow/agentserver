// UaaToken project main.go
package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	. "fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

func ListCFAPP(token string) *PaginatedApplicationResources {
	client := http.Client{}
	cfapiurl := "http://api." + conf.Cf_url
	path := Sprintf("%s/v2/apps", cfapiurl)
	req, _ := http.NewRequest("GET", path, nil)
	req.Header.Set("Authorization", "bearer "+token)
	req.Header.Set("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		Println(err.Error())
	}
	buff := new(bytes.Buffer)
	buff.ReadFrom(resp.Body)
	Println("%v", buff)
	Apps := new(PaginatedApplicationResources)
	_ = json.Unmarshal(buff.Bytes(), Apps)
	return Apps
}

func GetAppInstances(resource ApplicationResource, token string) int {
	cfapiurl := "http://api." + conf.Cf_url
	HttpClient := http.Client{}
	path := Sprintf("%s/v2/apps/"+resource.Metadata.Guid+"/instances", cfapiurl)
	HttpReq, _ := http.NewRequest("GET", path, nil)
	HttpReq.Header.Set("Authorization", "bearer "+token)
	HttpReq.Header.Set("Accept", "application/json")
	HttpResp, err := HttpClient.Do(HttpReq)
	if err != nil {
		Println(err.Error())
	}
	Buff := new(bytes.Buffer)
	Buff.ReadFrom(HttpResp.Body)
	Inst := make(map[string]map[string]interface{})
	err = json.Unmarshal(Buff.Bytes(), &Inst)
	if err != nil {
		//restart app
	}
	ans := 0
	for _, instance := range Inst {
		//Println(index, instance["state"])
		if instance["state"] == "RUNNING" {
			//restart app
			ans++
		}
	}
	return ans
}
func GetAuthToken(data url.Values) string {
	type uaaErrorResponse struct {
		Code        string `json:"error"`
		Description string `json:"error_description"`
	}
	type AuthenticationResponse struct {
		AccessToken  string           `json:"access_token"`
		TokenType    string           `json:"token_type"`
		RefreshToken string           `json:"refresh_token"`
		Error        uaaErrorResponse `json:"error"`
	}
	path := Sprintf("%s/oauth/token", "http://uaa."+conf.Cf_url)
	client := http.Client{}
	body := ioutil.NopCloser(strings.NewReader(data.Encode())) //把form数据编下码
	req, _ := http.NewRequest("POST", path, body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded;param=value")
	accessToken := "Basic " + base64.StdEncoding.EncodeToString([]byte("cf:"))
	req.Header.Set("Authorization", accessToken)
	resp, err := client.Do(req)
	if err != nil {
		Println(err.Error())
	}
	buff := new(bytes.Buffer)
	buff.ReadFrom(resp.Body)
	response := new(AuthenticationResponse)
	_ = json.Unmarshal(buff.Bytes(), &response)
	return response.AccessToken
}
