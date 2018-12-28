package main

/*
#cgo CFLAGS: -g -Wall
#cgo LDFLAGS: -lslurm

#include <slurm/slurm.h>
// The following is exported from src/common/slurm_protocol_defs.h
char *job_state_string(uint32_t inx)
{
	if (inx & JOB_COMPLETING)
		return "Completing";
	if (inx & JOB_STAGE_OUT)
		return "Stage_out";
	if (inx & JOB_CONFIGURING)
		return "Configuring";
	if (inx & JOB_RESIZING)
		return "Resizing";
	if (inx & JOB_REQUEUE)
		return "Requeued";
	if (inx & JOB_REQUEUE_FED)
		return "Requeue_fed";
	if (inx & JOB_REQUEUE_HOLD)
		return "Requeue_hold";
	if (inx & JOB_SPECIAL_EXIT)
		return "Special_exit";
	if (inx & JOB_STOPPED)
		return "Stopped";
	if (inx & JOB_REVOKED)
		return "Revoked";
	if (inx & JOB_RESV_DEL_HOLD)
		return "Resv_del_hold";
	if (inx & JOB_SIGNALING)
		return "Signaling";

	switch (inx & JOB_STATE_BASE) {
	case JOB_PENDING:
		return "Pending";
	case JOB_RUNNING:
		return "Running";
	case JOB_SUSPENDED:
		return "Suspended";
	case JOB_COMPLETE:
		return "Completed";
	case JOB_CANCELLED:
		return "Cancelled";
	case JOB_FAILED:
		return "Failed";
	case JOB_TIMEOUT:
		return "Timeout";
	case JOB_NODE_FAIL:
		return "Node_fail";
	case JOB_PREEMPTED:
		return "Preempted";
	case JOB_BOOT_FAIL:
		return "Boot_fail";
	case JOB_DEADLINE:
		return "Deadline";
	case JOB_OOM:
		return "Out_of_memory";
	default:
		return "?";
	}
}

char *job_state_string_compact(uint32_t inx)
{
	if (inx & JOB_COMPLETING)
		return "CG";
	if (inx & JOB_STAGE_OUT)
		return "SO";
	if (inx & JOB_CONFIGURING)
		return "CF";
	if (inx & JOB_RESIZING)
		return "RS";
	if (inx & JOB_REQUEUE)
		return "RQ";
	if (inx & JOB_REQUEUE_FED)
		return "RF";
	if (inx & JOB_REQUEUE_HOLD)
		return "RH";
	if (inx & JOB_SPECIAL_EXIT)
		return "SE";
	if (inx & JOB_STOPPED)
		return "ST";
	if (inx & JOB_REVOKED)
		return "RV";
	if (inx & JOB_RESV_DEL_HOLD)
		return "RD";
	if (inx & JOB_SIGNALING)
		return "SI";

	switch (inx & JOB_STATE_BASE) {
	case JOB_PENDING:
		return "PD";
	case JOB_RUNNING:
		return "R";
	case JOB_SUSPENDED:
		return "S";
	case JOB_COMPLETE:
		return "CD";
	case JOB_CANCELLED:
		return "CA";
	case JOB_FAILED:
		return "F";
	case JOB_TIMEOUT:
		return "TO";
	case JOB_NODE_FAIL:
		return "NF";
	case JOB_PREEMPTED:
		return "PR";
	case JOB_BOOT_FAIL:
		return "BF";
	case JOB_DEADLINE:
		return "DL";
	case JOB_OOM:
		return "OOM";
	default:
		return "?";
	}
}
*/
import "C"

import (
	"fmt"
	"github.com/buger/jsonparser"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unsafe"
)

// Structure Definition (following format needed by `hqueue`, `jnode`)
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
	JobID     string   `json:"jobid"`
	Status    string   `json:"status"`
	User      string   `json:"user"`
	Jobname   string   `json:"jobname"`
	Node      []string `json:"node"`
	Elapsed   string   `json:"elapsed"`
	GPU       JobGPU   `json:"gpu"`
	CPU       JobCPU   `json:"cpu"`
	Billing   int64    `json:"billing"`
	Spot      bool     `json:"spot"`
	Remaining string   `json:"remaining"`
}

// Utility Function
func (info *JobInfoType) setTres(tres string) {
	//example:cpu=8,mem=20G,node=1,billing=88
	for _, tre := range strings.Split(tres, ",") {
		switch s := strings.Split(tre, "="); s[0] {
		case "cpu":
			cpuAlloc, _ := strconv.Atoi(s[1])
			info.CPU.Core = map[string]int64{"total": int64(cpuAlloc), "util": 0}
		case "mem":
			memAlloc, _ := strconv.Atoi(s[1][:len(s[1])-1])
			switch unit := string(s[1][len(s[1])-1]); unit {
			case "T":
				memAlloc <<= 20
			case "G":
				memAlloc <<= 10
			case "K":
				memAlloc >>= 10
			}
			info.CPU.MEM = map[string]int64{"total": int64(memAlloc), "util": 0}
		case "billing":
			billing, _ := strconv.Atoi(s[1])
			info.Billing = int64(billing)
		}
	}
}

func parseJobGres(gres string) []int { //gpu(IDX:0-2), gpu(IDX:0)
	idxAllS := strings.Split(gres, "IDX:")[1]
	idxAllS = idxAllS[:len(idxAllS)-1]
	// e.g 0-2,3,5-7
	return parseGresIdx(idxAllS)

}

func (info *JobInfoType) init(ndJSON []uint8, slurmInfo _Ctype_struct_job_info, nodePtrMap map[string]*NodeInfoType, now time.Time) {
	info.JobID = strconv.FormatUint(uint64(slurmInfo.job_id), 10)
	uid := strconv.FormatUint(uint64(slurmInfo.user_id), 10)
	info.Status = C.GoString(C.job_state_string(slurmInfo.job_state))
	info.setTres(C.GoString(slurmInfo.tres_alloc_str))
	info.User = C.GoString(slurmInfo.account)
	info.Jobname = C.GoString(slurmInfo.name)
	var startTime = time.Unix(int64(slurmInfo.start_time), 0)
	info.Elapsed = now.Sub(startTime).Round(time.Second).String()
	if C.GoString(slurmInfo.qos) == "spot" {
		info.Spot = true
	}
	info.Remaining = time2str(int64(slurmInfo.time_limit))

	if ndJSON != nil {
		//FIXME: brute force enumerate possible usage here, should use ObjectEach to summation over values
		cpuSysUtil, _ := jsonparser.GetFloat(ndJSON, fmt.Sprintf(JOB_USAGE, uid, info.JobID, "cpu"), "dimensions", "system", "value")
		cpuUserUtil, _ := jsonparser.GetFloat(ndJSON, fmt.Sprintf(JOB_USAGE, uid, info.JobID, "cpu"), "dimensions", "user", "value")
		info.CPU.Core["util"] = int64(cpuSysUtil) + int64(cpuUserUtil)

		memUtil, _ := jsonparser.GetFloat(ndJSON, fmt.Sprintf(JOB_USAGE, uid, info.JobID, "mem_usage"), "dimensions", "ram", "value")
		info.CPU.MEM["util"] = int64(memUtil)
	}

	//TODO: modify to list later? (to support "single job on multi node")
	//1. ndJSON should be a list
	//2. How to show CPU/MEM stat (display individually or summation over them?)
	info.Node = append(info.Node, C.GoString(slurmInfo.nodes)) // assume single node

	// Set job info to that occupied node
	//TODO: We don't know the mapping between allocated nodes and gres_detail...
	//      Since gres_cnt at most 1 at the moment, just hardcode here
	gresCnt := int(slurmInfo.gres_detail_cnt)
	gresArr := *(*[]*C.char)(unsafe.Pointer(&reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(slurmInfo.gres_detail_str)),
		Len:  gresCnt,
		Cap:  gresCnt,
	}))

	for _, node := range info.Node {
		if occupyNode, ok := nodePtrMap[node]; ok {
			gpus := []int{}
			if gresCnt == 1 { // if that job use GPU, need to modify if use multi node
				gpus = parseJobGres(C.GoString(gresArr[0]))
			}

			occupyNode.Jobs = append(occupyNode.Jobs, NodeJobType{
				JobID: info.JobID,
				User:  info.User,
				GPUs:  gpus,
				Spot:  info.Spot,
			})

			for _, gpu := range gpus {
				if info.GPU.MEM == nil {
					info.GPU.MEM = map[string]int64{"total": 0, "util": 0}
				}
				idxTuple := []interface{}{node, gpu}
				info.GPU.Idxs = append(info.GPU.Idxs, idxTuple)
				info.GPU.Models = append(info.GPU.Models, occupyNode.GPUs[gpu].Model)
				info.GPU.Util += int64(occupyNode.GPUs[gpu].Util)
				info.GPU.MEM["total"] += occupyNode.GPUs[gpu].MemTotal
				info.GPU.MEM["util"] += int64(occupyNode.GPUs[gpu].MEM)
			}
		}
	}
}
