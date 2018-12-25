package main

/*
#cgo CFLAGS: -g -Wall
#cgo LDFLAGS: -lslurm

#include <slurm/slurm.h>
*/
import "C"
import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"
	//"os"
	"unsafe"
)

func main() {
	//var period = flag.Float64("t", 1.0, "update info every t seconds")
	var port = flag.Int("port", 19999, "port")

	var sNodeInfoMgr *C.node_info_msg_t
	C.slurm_load_node(0, &sNodeInfoMgr, C.SHOW_DETAIL)
	sData := unsafe.Pointer(sNodeInfoMgr.node_array)
	numNodes := int(sNodeInfoMgr.record_count)

	nodeArr := *(*[]C.node_info_t)(unsafe.Pointer(&reflect.SliceHeader{
		Data: uintptr(sData),
		Len:  numNodes,
		Cap:  numNodes,
	}))

	for i := 0; i < numNodes; i++ {
		hostname := C.GoString(nodeArr[i].node_hostname)
		if hostname != "s01" {
			continue
		}
		instance := fmt.Sprintf(INSTANCE, *port, hostname, DOMAIN, API_VERSION)
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
			//fmt.Printf("%s\n", resp_data)
			continue
		}

		//foo, err := jsonparser.GetInt(resp_data, "sensors.coretemp-isa-0000_temperature", "dimensions", "coretemp-isa-0000_temp1", "nae")
		//fmt.Println(err)
		//fmt.Println(foo)
		//fmt.Println("ffdasfs")
		var nodeInfo NodeInfoType
		nodeInfo.init(resp_data, nodeArr[i])
		test_data, _ := json.Marshal(nodeInfo)
		fmt.Printf("%s\n", test_data)
	}
	// no need defer, directly release
	C.slurm_free_node_info_msg(sNodeInfoMgr)

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
