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
	"time"
	//"os"
	"unsafe"
)

func main() {
	//var period = flag.Float64("t", 1.0, "update info every t seconds")
	timeBaseline := time.Now()
	var port = flag.Int("port", 19999, "port")
	var endpoint = flag.String("e", "N/A", "endpoint [node/job]")

	// NODE info
	nodeAccessMap := make(map[string]*NodeInfoType)
	nodeJSONMap := make(map[string][]uint8)

	var nodeList []NodeInfoType
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

		//TODO: remove it
		if len(resp_data) < 100 { // no netdata
			//fmt.Printf("%s\n", resp_data)
			continue
		}

		var nodeInfo NodeInfoType
		nodeInfo.init(resp_data, nodeArr[i])
		nodeList = append(nodeList, nodeInfo)
		nodeJSONMap[nodeInfo.Hostname] = resp_data
		defer resp.Body.Close()
	}
	// Build NodeInfoAccess Map (Note that need to do AFTER ALL NODE APPEND TO
	// NODELIST), since append will cause reallocate, and the pointer may not be
	// valid (even not crash...= =) ,and the last will correct
	for i, nd := range nodeList {
		nodeAccessMap[nd.Hostname] = &(nodeList[i])
	}
	C.slurm_free_node_info_msg(sNodeInfoMgr)

	var jobList []JobInfoType
	var sJobInfoMgr *C.job_info_msg_t
	C.slurm_load_jobs(0, &sJobInfoMgr, C.SHOW_DETAIL)
	sJobData := unsafe.Pointer(sJobInfoMgr.job_array)
	numJobs := int(sJobInfoMgr.record_count)

	jobArr := *(*[]C.slurm_job_info_t)(unsafe.Pointer(&reflect.SliceHeader{
		Data: uintptr(sJobData),
		Len:  numJobs,
		Cap:  numJobs,
	}))

	for i := 0; i < numJobs; i++ {
		var jobInfo JobInfoType
		//TODO: hostname may not be only one node
		hostname := C.GoString(jobArr[i].nodes)
		jobInfo.init(nodeJSONMap[hostname], jobArr[i], nodeAccessMap, timeBaseline)
		test_data, _ := json.Marshal(jobInfo)
		fmt.Printf("%s\n", test_data)
		//jobList = append(jobList, jobInfo)
		//break
		fmt.Println()
	}

	for _, v := range nodeList {
		test_data, _ := json.Marshal(v)
		fmt.Printf("%s\n", test_data)
	}

	C.slurm_free_job_info_msg(sJobInfoMgr)
}
