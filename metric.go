package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

type TSDB struct {
	host string
	port string
}

type Response []result

type Timestamp struct {
	time.Time
}

type Metric struct {
	string
}

type Value struct {
	float64
}

type Tags struct {
	tags map[string]string
}

type result struct {
	Metric         string             `json:"metric"`
	Tags           map[string]string  `json:"tags"`
	AggregatedTags []string           `json:"aggregateTags"`
	Dps            map[string]float64 `json:"dps"`
}

/* tsdb request json format
{
    "start": 1356998400,
    "end": 1356998460,
    "queries": [
        {
            "aggregator": "sum",
            "metric": "sys.cpu.0",
            "rate": "true",
            "tags": {
                "host": "*",
                "dc": "lga"
            }
        },
        {
            "aggregator": "sum",
            "tsuids": [
                "000001000002000042",
                "000001000002000043"
              ]
            }
        }
    ]
}
*/

/* tsdb response json format

[
    {
        "metric": "tsd.hbase.puts",
        "tags": {},
        "aggregatedTags": [
            "host"
        ],
        "dps": {
            "1365966001": 25595461080,
            "1365966061": 25595542522,
            "1365966062": 25595543979,
...
            "1365973801": 25717417859
        }
    }
]
*/

type Query struct {
	Aggregator string            `json:"aggregator"`
	Metric     string            `json:"metric"`
	Rate       bool              `json:"rate"`
	Tags       map[string]string `json:"tags"`
}

type Time struct {
	time   time.Time
	format string
	string
}

func (t *Time) MarshalJSON() ([]byte, error) {
	switch t.format {
	case "":
		return nil, nil
	case "Relative":
		return json.Marshal(t.string)
	default:
		return json.Marshal(t.time.Unix())
	}
}

type Request struct {
	Start   *Time   `json:"start"`
	Queries []Query `json:"queries"`
}

func (m *TSDB) Get(metric string, aggregator string, relative_time string, job string, index int) (Response, error) {
	mp := map[string]string{}
	mp["job"] = job
	mp["index"] = strconv.Itoa(index)

	que := Query{
		Aggregator: aggregator,
		Metric:     metric,
		Rate:       false,
		Tags:       mp,
	}

	tm := Time{}
	tm.format = "Relative"
	tm.string = relative_time

	reqStr := Request{
		Start:   &tm,
		Queries: []Query{que},
	}

	reqJSON, err := json.Marshal(reqStr)
	reqReader := bytes.NewReader(reqJSON)

	host := m.host + ":" + m.port
	APIURL := "http://" + host + "/api/query"

	c := &http.Client{}

	req, _ := http.NewRequest("POST", APIURL, reqReader)
	resp, err := c.Do(req)

	if err != nil {
		fmt.Printf("Request tsdb metric error: ip %s %s\n", m.host, m.port)
		return Response{}, err
	}
	if resp.StatusCode != 200 {
		fmt.Printf("Request tsdb error with response code: %v\n", resp.StatusCode)
		fmt.Println("Requst info: ", reqJSON)
		return Response{}, err
	}

	respJSON, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("error: %v", err)
		return Response{}, err
	}

	ans := Response{}
	err = json.Unmarshal(respJSON, &ans)
	if err != nil {
		fmt.Println("error: %v", err)
		return Response{}, err
	}
	return ans, nil
}

func (m *TSDB) ListMetrics(maxnum int) ([]string, error) {
	c := &http.Client{}
	//http://10.10.101.146:4242/api/suggest?type=metrics&max=200
	host := m.host + ":" + m.port
	APIURL := "http://" + host + "/api/suggest?type=metrics&max=" + strconv.Itoa(maxnum)
	req, _ := http.NewRequest("GET", APIURL, nil)
	resp, err := c.Do(req)
	ans := []string{}
	respJSON, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("error: %v", err)
		return ans, err
	}
	json.Unmarshal(respJSON, &ans)
	return ans, nil
}

func (m *TSDB) ListMetricsWithQ(maxnum int, q string) ([]string, error) {
	c := &http.Client{}
	host := m.host + ":" + m.port
	APIURL := "http://" + host + "/api/suggest?type=metrics&q=" + q + "&max=" + strconv.Itoa(maxnum)
	req, _ := http.NewRequest("GET", APIURL, nil)
	resp, err := c.Do(req)
	ans := []string{}
	respJSON, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("error: %v", err)
		return ans, err
	}
	json.Unmarshal(respJSON, &ans)
	return ans, nil
}
