package main

/*
#cgo CFLAGS: -g -Wall
#cgo LDFLAGS: -lslurm

#include <slurm/slurm.h>
*/
import "C"
import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"
	"time"
	//"os"
	"unsafe"
)

var prepared bool
var nodeAccessMap map[string]*NodeInfoType
var nodeJSONMap map[string][]uint8
var nodeList []NodeInfoType
var jobList []JobInfoType

func reset() {
	nodeJSONMap = nil
	nodeAccessMap = nil
	nodeList = nil
	jobList = nil
}

func pollInfo() {
	// approx. 50-100 ms
	//TODO: How to lock this
	for {
		//s := time.Now()
		prepared = false
		reset()
		timeBaseline := time.Now()
		setNodeInfo()
		// Build NodeInfoAccess Map (Note that need to do AFTER ALL NODE APPEND TO
		// NODELIST), since append will cause reallocate, and the pointer may not be
		// valid (even not crash...= =) ,and the last will correct
		for i, node := range nodeList {
			nodeAccessMap[node.Hostname] = &(nodeList[i])
		}
		setJobInfo(timeBaseline)
		prepared = true
		//e := time.Now()
		//fmt.Printf("%s", e.Sub(s).String())

		time.Sleep(DELAY * time.Second)
	}
}

func setNodeInfo() {

	nodeAccessMap = make(map[string]*NodeInfoType)
	nodeJSONMap = make(map[string][]uint8)

	var sNodeInfoMgr *C.node_info_msg_t
	C.slurm_load_node(0, &sNodeInfoMgr, C.SHOW_DETAIL)
	sData := unsafe.Pointer(sNodeInfoMgr.node_array)
	numNodes := int(sNodeInfoMgr.record_count)

	nodeArr := *(*[]C.node_info_t)(unsafe.Pointer(&reflect.SliceHeader{
		Data: uintptr(sData),
		Len:  numNodes,
		Cap:  numNodes,
	}))

	//TODO: can use goroutine to parallel query
	for i := 0; i < numNodes; i++ {
		hostname := C.GoString(nodeArr[i].node_hostname)

		instance := fmt.Sprintf(INSTANCE, PORT, hostname, DOMAIN, API_VERSION)
		infoQuery := fmt.Sprintf("%s/%s", instance, ALL_METRIC_POINT)

		// Get netdata
		resp, err := http.Get(infoQuery)
		if err != nil {
			log.Fatal(err)
		}
		respData, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}

		//TODO: remove it
		if len(respData) < 100 { // no netdata
			//fmt.Printf("%s\n", respData)
			continue
		}

		var nodeInfo NodeInfoType
		nodeInfo.init(respData, nodeArr[i])
		nodeList = append(nodeList, nodeInfo)
		nodeJSONMap[nodeInfo.Hostname] = respData
		resp.Body.Close()
	}
	C.slurm_free_node_info_msg(sNodeInfoMgr)
}

func setJobInfo(now time.Time) {
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
		jobInfo.init(nodeJSONMap[hostname], jobArr[i], nodeAccessMap, now)
		jobList = append(jobList, jobInfo)
	}
	C.slurm_free_job_info_msg(sJobInfoMgr)
}

func queryNodeInfo(w http.ResponseWriter, r *http.Request) {
	for !prepared {
		fmt.Println("Data haven't been prepared yet")
		time.Sleep(100 * time.Millisecond)
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-type", "application/json")
	data := struct {
		Success   bool           `json:"success"`
		NodeInfos []NodeInfoType `json:"nodeInfos"`
	}{true, nodeList}
	res, _ := json.Marshal(data)
	w.Write(res)
}

func queryJobInfo(w http.ResponseWriter, r *http.Request) {
	for !prepared {
		fmt.Println("Data haven't been prepared yet")
		time.Sleep(100 * time.Millisecond)
	}
	w.WriteHeader(http.StatusOK)
	user := r.Header.Get("x-user")
	w.Header().Set("Content-type", "application/json")
	userJobList := make([]JobInfoType, 0)
	for _, job := range jobList {
		if job.User == user {
			userJobList = append(userJobList, job)
		}
	}
	data := struct {
		Success bool          `json:"success"`
		JobInfo []JobInfoType `json:"jobInfo"`
	}{true, userJobList}
	res, _ := json.Marshal(data)
	w.Write(res)
}

func main() {
	go pollInfo()

	http.HandleFunc("/monitor/nodes", queryNodeInfo)
	http.HandleFunc("/monitor/jobs", queryJobInfo)
	//TODO: add not found handler? (maybe no needed)
	err := http.ListenAndServe("localhost:8888", nil)

	if err != nil {
		log.Fatal("Server Error", err)
	}

}
