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
	"sync"
	"time"
	"unsafe"
)

//var prepared bool
var nodeAccessMap map[string]*NodeInfoType

var nodeJSONMap map[string][]uint8

//var nodeJSONMap sync.Map
var nodeList []NodeInfoType
var jobList []JobInfoType
var mu sync.RWMutex

func reset() {
	//nodeJSONMap.Range(func(key interface{}, value interface{}) bool {
	//nodeJSONMap.Delete(key)
	//return true
	//})
	nodeJSONMap = nil
	nodeAccessMap = nil
	nodeList = nil
	jobList = nil
}

func pollInfo() {
	for {
		s := time.Now()
		mu.Lock()
		//prepared = false
		reset()
		setNodeInfo()
		// Build NodeInfoAccess Map (Note that need to do AFTER ALL NODE APPEND TO
		// NODELIST), since append will cause reallocate, and the pointer may not be
		// valid (even not crash...= =) ,and the last will correct
		nodeAccessMap = make(map[string]*NodeInfoType)
		for i, node := range nodeList {
			nodeAccessMap[node.Hostname] = &(nodeList[i])
		}
		setJobInfo()
		//prepared = true
		mu.Unlock()
		e := time.Now()
		fmt.Printf("%s\n", e.Sub(s).String())

		time.Sleep(DELAY * time.Second)
		//time.Sleep(100 * time.Millisecond)
	}
}

func setNodeInfo() {

	var sNodeInfoMgr *C.node_info_msg_t
	C.slurm_load_node(0, &sNodeInfoMgr, C.SHOW_DETAIL)
	sData := unsafe.Pointer(sNodeInfoMgr.node_array)
	numNodes := int(sNodeInfoMgr.record_count)

	nodeArr := *(*[]C.node_info_t)(unsafe.Pointer(&reflect.SliceHeader{
		Data: uintptr(sData),
		Len:  numNodes,
		Cap:  numNodes,
	}))

	//nodeJSONMap = make(map[string][]uint8)
	//for i := 0; i < numNodes; i++ {
	//nodeJSONMap[C.GoString(nodeArr[i].node_hostname)] = make([]uint8, 0)
	//}

	//Parallel
	//nodeListTmp := make([]NodeInfoType, numNodes)
	//wg := sync.WaitGroup{}

	//TODO: since concurrent write to map will cause race issue
	//and the solution using sync.Map seems not help the efficiency, give up
	//parallel here
	for i := 0; i < numNodes; i++ {
		//wg.Add(1)
		//go func(i int) {
		hostname := C.GoString(nodeArr[i].node_hostname)
		//fmt.Println(hostname)

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

		//nodeJSONMap.Store(hostname, respData)
		nodeJSONMap = make(map[string][]uint8, 0)
		nodeJSONMap[hostname] = respData

		//TODO: remove it
		if len(respData) > 100 { // no netdata
			var nodeInfo NodeInfoType
			nodeInfo.init(respData, nodeArr[i])
			nodeList = append(nodeList, nodeInfo)
		}
		resp.Body.Close()
		//defer wg.Done()
		//}(n)
	}
	//wg.Wait()
	//TODO: reallocate (should remove later?)
	//for _, nodeInfo := range nodeListTmp {
	//nodeList = append(nodeList, nodeInfo)
	//}
	defer C.slurm_free_node_info_msg(sNodeInfoMgr)
}

func setJobInfo() {
	var now = time.Now()
	var sJobInfoMgr *C.job_info_msg_t
	C.slurm_load_jobs(0, &sJobInfoMgr, C.SHOW_DETAIL)
	sJobData := unsafe.Pointer(sJobInfoMgr.job_array)
	numJobs := int(sJobInfoMgr.record_count)

	jobArr := *(*[]C.slurm_job_info_t)(unsafe.Pointer(&reflect.SliceHeader{
		Data: uintptr(sJobData),
		Len:  numJobs,
		Cap:  numJobs,
	}))

	// CANNOT parallel (since job init() include some operate on NodeInfo)
	jobList = make([]JobInfoType, numJobs)

	for j := 0; j < numJobs; j++ {
		var jobInfo JobInfoType
		//TODO: hostname may not be only one node
		hostname := C.GoString(jobArr[j].nodes)
		//var jsonMap []uint8
		//jsonMap, _ = nodeJSONMap.Load(hostname)
		jobInfo.init(nodeJSONMap[hostname], jobArr[j], nodeAccessMap, now)
		jobList[j] = jobInfo
	}

	defer C.slurm_free_job_info_msg(sJobInfoMgr)
}

func queryNodeInfo(w http.ResponseWriter, r *http.Request) {
	mu.RLock()
	//for !prepared {
	//fmt.Println("Data haven't been prepared yet")
	//time.Sleep(100 * time.Millisecond)
	//}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-type", "application/json")
	data := struct {
		Success   bool           `json:"success"`
		NodeInfos []NodeInfoType `json:"nodeInfos"`
	}{true, nodeList}
	res, _ := json.Marshal(data)
	w.Write(res)
	defer mu.RUnlock()
}

func queryJobInfo(w http.ResponseWriter, r *http.Request) {
	mu.RLock()
	//for !prepared {
	//fmt.Println("Data haven't been prepared yet")
	//time.Sleep(100 * time.Millisecond)
	//}
	w.WriteHeader(http.StatusOK)
	user := r.Header.Get("x-user")
	w.Header().Set("Content-type", "application/json")
	userJobList := make([]JobInfoType, 0)
	for _, job := range jobList {
		if job.User == user {
			userJobList = append(userJobList, job)
		}
	}
	if len(userJobList) == 0 {
		userJobList = jobList
	}
	data := struct {
		Success bool          `json:"success"`
		JobInfo []JobInfoType `json:"jobInfo"`
	}{true, userJobList}
	res, _ := json.Marshal(data)
	w.Write(res)
	defer mu.RUnlock()
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
