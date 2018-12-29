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

var nodeAccessMap map[string]*NodeInfoType

var nodeJSONMap map[string][]uint8
var nodeList []NodeInfoType
var jobList []JobInfoType
var mu sync.RWMutex

type JSONdata struct {
	hostname string
	data     []uint8
}

func reset() {
	nodeJSONMap = nil
	nodeAccessMap = nil
	nodeList = nil
	jobList = nil
}

func pollInfo() {
	for {
		s := time.Now()
		mu.Lock()
		reset()
		setNodeInfo()
		nodeAccessMap = make(map[string]*NodeInfoType)
		for i, node := range nodeList {
			nodeAccessMap[node.Hostname] = &(nodeList[i])
		}
		setJobInfo()
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
	nodeJSONMap = make(map[string][]uint8, 0)
	JSONcache := make(chan JSONdata, numNodes)

	nodeArr := *(*[]C.node_info_t)(unsafe.Pointer(&reflect.SliceHeader{
		Data: uintptr(sData),
		Len:  numNodes,
		Cap:  numNodes,
	}))

	for i := 0; i < numNodes; i++ {
		go func(hostname string) {
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
			JSONcache <- JSONdata{hostname, respData}
			defer resp.Body.Close()
		}(C.GoString(nodeArr[i].node_hostname))
	}

	//TODO: should await here?
	// Dump channel
	// since map is not thread-safe (even we have unique key when inserting)
	// see `go build -race`
	for i := 0; i < numNodes; i++ {
		tmp := <-JSONcache
		//TODO: remove it
		if len(tmp.data) > 100 {
			nodeJSONMap[tmp.hostname] = tmp.data
		}
	}
	fmt.Println(len(JSONcache)) // 0

	for i := 0; i < numNodes; i++ {
		hostname := C.GoString(nodeArr[i].node_hostname)
		if ndData, ok := nodeJSONMap[hostname]; ok {
			var nodeInfo NodeInfoType
			nodeInfo.init(ndData, nodeArr[i])
			nodeList = append(nodeList, nodeInfo)
		}
	}
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
		jobInfo.init(nodeJSONMap[hostname], jobArr[j], nodeAccessMap, now)
		jobList[j] = jobInfo
	}
	defer C.slurm_free_job_info_msg(sJobInfoMgr)
}

func queryNodeInfo(w http.ResponseWriter, r *http.Request) {
	mu.RLock()
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
