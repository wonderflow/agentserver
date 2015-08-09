package main

import (
	"fmt"
	"strings"
)

func Host_alive_check() []string {
	var (
		ips         []string
		ips_removed []string
		ans         []string
	)
	response, err := client.Get(conf.Etcd_dir, true, true)
	if err != nil {
		fmt.Printf("Etcd get alive node error: %v\n", err)
		return ans
	}
	for i := 0; i < len(response.Node.Nodes); i++ {
		key := response.Node.Nodes[i].Key
		ip := strings.Split(key, "/")[2]
		ips = append(ips, ip)
	}

	response, err = client.Get(conf.Etcd_rm_dir, true, true)

	if err != nil {
		if strings.Contains(err.Error(), "Key not found") != true {
			fmt.Println(err)
		}
	} else {
		for i := 0; i < len(response.Node.Nodes); i++ {
			ip := response.Node.Nodes[i].Value
			ips_removed = append(ips_removed, ip)
		}
	}

	for i := 0; i < len(ips_removed); i++ {
		fmt.Printf("Detected Host removed: %v\n", ips_removed[i])
		for j := 0; j < len(ips); j++ {
			if ips[j] == ips_removed[i] {
				for k := j; k < len(ips)-1; k++ {
					ips[k] = ips[k+1]
				}
				ips = ips[:len(ips)-1]
				break
			}
		}
		for j := 0; j < len(should_alived_host); j++ {
			if should_alived_host[j] == ips_removed[i] {
				for k := j; k < len(should_alived_host)-1; k++ {
					should_alived_host[k] = should_alived_host[k+1]
				}
				should_alived_host = should_alived_host[:len(should_alived_host)-1]
				break
			}
		}
	}
	if len(ips_removed) > 0 {
		fmt.Printf("After removed IP should alive in host %v\n", should_alived_host)
	}

	l1 := len(should_alived_host)
	l2 := len(ips)

	for i := 0; i < l2; i++ {
		j := 0
		for ; j < l1; j++ {
			if ips[i] == should_alived_host[j] {
				break
			}
		}
		if j == l1 {
			should_alived_host = append(should_alived_host, ips[i])
			fmt.Printf("HOST %v detected and added.\n", ips[i])
		}
	}

	l1 = len(should_alived_host)
	l2 = len(ips)

	for i := 0; i < l1; i++ {
		j := 0
		for ; j < l2; j++ {
			if should_alived_host[i] == ips[j] {
				break
			}
		}
		if j == l2 {
			ans = append(ans, should_alived_host[i])
			fmt.Printf("HOST %v missed.\n", should_alived_host[i])
		}
	}
	return ans

}
