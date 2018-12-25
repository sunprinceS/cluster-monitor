package main

import (
	"encoding/json"
	"flag"
	"fmt"
	//"github.com/buger/jsonparser"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

func main() {
	//var period = flag.Float64("t", 1.0, "update info every t seconds")
	var port = flag.Int("port", 9999, "port")

	instance := fmt.Sprintf(INSTANCE, *port, "s01", DOMAIN, API_VERSION)
	info_query := fmt.Sprintf("%s/%s", instance, ALL_METRIC_POINT)
	flag.Parse()

	// Get netdata
	resp, err := http.Get(info_query)
	if err != nil {
		log.Fatal(err)
	}
	resp_data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	//TODO: remove it
	if len(resp_data) < 100 { // no netdata
		os.Exit(1)
	}

	var nodeInfo NodeInfo
	nodeInfo.init(resp_data)
	test_data, _ := json.Marshal(nodeInfo)
	fmt.Printf("%s\n", test_data)

	//if resp_data
	//fmt.Printf("%d", len(resp_data))
	//var nodeInfoList []NodeInfo
	//var foo_gpus []GPU
	//foo_gpus = append(foo_gpus,
	//GPU{
	//Util:      0,
	//Mem:       12300,
	//Model:     "2080Ti",
	//Mem_total: 21,
	//Used:      false,
	//Temp:      123,
	//},
	//GPU{
	//Util:      0,
	//Mem:       123,
	//Model:     "2080Ti",
	//Mem_total: 1,
	//Used:      false,
	//Temp:      1,
	//},
	//)
	//nodeInfoList = append(nodeInfoList,
	//NodeInfo{
	////nodeInfo := NodeInfo{
	//Hostname: "s01.speech",
	//Online:   true,
	//Normal:   true,
	//Cpu: CPU{
	//Util:  75.33,
	//Temp:  123,
	//Total: 321,
	//Alloc: 21,
	//},
	//Gpus: foo_gpus,
	//Mem: MEM{
	//Total: 32,
	//Alloc: 12,
	//Util:  22,
	//},
	//})

	//data, _ := json.Marshal(nodeInfoList)
	//fmt.Printf("%s\n", data)
}
