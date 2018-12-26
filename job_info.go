package main

/*
#cgo CFLAGS: -g -Wall
#cgo LDFLAGS: -lslurm

#include <slurm/slurm.h>
*/
import "C"
import (
	//"fmt"
	"time"
	//"github.com/buger/jsonparser"
	//"reflect"
	"strconv"
	"strings"
	//"unsafe"
)

type JobGPU struct {
	Idxs   []interface{}    `json:"idxs"`
	Models []string         `json:"models"`
	Util   int64            `json:"util"`
	MEM    map[string]int64 `json:"mem"`
}

type JobCPU struct {
	Core map[string]int64 `json:"core"`
	MEM  map[string]int64 `json:"mem"`
}

type JobInfoType struct {
	JobID   string   `json:"jobid"`
	Status  string   `json:"status"`
	User    string   `json:"user"`
	Jobname string   `json:"jobname"`
	Node    []string `json:"node"`
	Elapsed string   `json:"elapsed"`
	GPU     JobGPU   `json:"gpu"`
	CPU     JobCPU   `json:"cpu"`
	Billing int64    `json:"billing"`
	//FIXME:no dynamic data member in golang?
	Spot      bool   `json:"spot"`
	Remaining string `json:"remaining"`
}

func (info *JobInfoType) setState(jobState uint) {
	switch jobState {
	case 0:
		info.Status = "Pending"
	case 1:
		info.Status = "Running"
	case 2:
		info.Status = "Suspended"
	case 3:
		info.Status = "Complete"
	case 4:
		info.Status = "Cancelled"
	case 5:
		info.Status = "Failed"
	case 6:
		info.Status = "Timeout"
	default:
		info.Status = "OMIMI"
	}
}

//TODO
func (info *JobInfoType) setTres(tres string) { //alloctres cpu=8 mem=20G node=1 billing=88
	//tresStrLs := strings.Split(tres, ",")
	for _, tre := range strings.Split(tres, ",") {
		switch s := strings.Split(tre, "="); s[0] {
		case "cpu":
			cpuUsage, _ := strconv.Atoi(s[1])
			info.CPU.Core = map[string]int64{"total": int64(cpuUsage), "util": 0}
		case "mem":
			memUsageMB, _ := strconv.Atoi(s[1][:len(s[1])-1])
			switch unit := string(s[1][len(s[1])-1]); unit {
			case "T":
				memUsageMB <<= 20
			case "G":
				memUsageMB <<= 10
			case "K":
				memUsageMB >>= 10
			}
			info.CPU.MEM = map[string]int64{"total": int64(memUsageMB), "util": 0}
		case "billing":
			billing, _ := strconv.Atoi(s[1])
			info.Billing = int64(billing)
		}

	}
}

//func formatTime(t time.Time){
//}

func time2str(mins int64) string {
	if mins == int64(C.INFINITE) {
		return "INFINITE"
	} else {
		// unit: min
		return (time.Duration(mins) * time.Minute).String()
	}

}

//TODO:
//1. set cpu, mem util from netdataJSON cGroup
//2. set gpu usage from netdataJSON corresponding GPU
//3. parse Gres
func (info *JobInfoType) init(netdataJSON []uint8, sData _Ctype_struct_job_info, now time.Time) {
	info.JobID = strconv.FormatUint(uint64(sData.job_id), 10)
	info.setTres(C.GoString(sData.tres_alloc_str))
	info.User = C.GoString(sData.account)
	info.Jobname = C.GoString(sData.name)
	//TODO: modify to list later? (multi-node single job)
	info.Node = append(info.Node, C.GoString(sData.nodes))
	var startTime = time.Unix(int64(sData.start_time), 0)
	info.Elapsed = now.Sub(startTime).Round(time.Second).String()
	if C.GoString(sData.qos) == "spot" {
		info.Spot = true
	}
	info.Remaining = time2str(int64(sData.time_limit))
	info.setState(uint(sData.job_state))

	//OK!
	//gres_cnt := int(sData.gres_detail_cnt)
	//gres_detail_arr := *(*[]*C.char)(unsafe.Pointer(&reflect.SliceHeader{
	//Data: uintptr(unsafe.Pointer(sData.gres_detail_str)),
	//Len:  gres_cnt,
	//Cap:  gres_cnt,
	//}))
}
