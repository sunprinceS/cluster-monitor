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

func (info *JobInfoType) setTres(tres string) { //alloctres cpu=8 mem=20G node=1 billing=88
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
		return "UNLIMITED"
	} else {
		// unit: min
		return (time.Duration(mins) * time.Minute).String()
	}

}

func parseJobGres(gresStr string) (ret []int) { //gpu(IDX:0-2), gpu(IDX:0)
	gresIDXallStr := strings.Split(gresStr, "IDX:")[1]
	gresIDXallStr = gresIDXallStr[:len(gresIDXallStr)-1]
	gresIDXstrList := strings.Split(gresIDXallStr, ",")
	for _, gresIDXstr := range gresIDXstrList {
		gresIDXlist := strings.Split(gresIDXstr, "-")
		if len(gresIDXlist) == 1 {
			gres, _ := strconv.Atoi(gresIDXlist[0])
			ret = append(ret, gres)
		} else {
			start, _ := strconv.Atoi(gresIDXlist[0])
			end, _ := strconv.Atoi(gresIDXlist[1])
			for gres := start; gres <= end; gres++ {
				ret = append(ret, gres)
			}
		}
	}
	return

	//info.Idxs := []interface{}
	//info.GPU.Idxs = append(info.GPU.Idxs, 1, "s01")
	//fmt.Println("Hwllo Wor")
}

//TODO:
//2. set gpu usage from netdataJSON corresponding GPU
//3. parse Gres (then set corresponding GPU idx in node_info.gpus)
func (info *JobInfoType) init(netdataJSON []uint8, sData _Ctype_struct_job_info, nodeAccessMap map[string]*NodeInfoType, now time.Time) {
	info.JobID = strconv.FormatUint(uint64(sData.job_id), 10)
	UID := strconv.FormatUint(uint64(sData.user_id), 10)
	info.Status = C.GoString(C.job_state_string(sData.job_state))
	info.setTres(C.GoString(sData.tres_alloc_str))
	info.User = C.GoString(sData.account)
	info.Jobname = C.GoString(sData.name)
	var startTime = time.Unix(int64(sData.start_time), 0)
	info.Elapsed = now.Sub(startTime).Round(time.Second).String()
	if C.GoString(sData.qos) == "spot" {
		info.Spot = true
	}
	info.Remaining = time2str(int64(sData.time_limit))

	if netdataJSON != nil {
		memUtil, _ := jsonparser.GetFloat(netdataJSON, fmt.Sprintf(JOB_USAGE, UID, info.JobID, "mem_usage"), "dimensions", "ram", "value")
		info.CPU.MEM["util"] = int64(memUtil)
		//FIXME: brute force enumerate here :p, using ObjectEach to sum it?
		cpuSysUtil, _ := jsonparser.GetFloat(netdataJSON, fmt.Sprintf(JOB_USAGE, UID, info.JobID, "cpu"), "dimensions", "system", "value")
		cpuUserUtil, _ := jsonparser.GetFloat(netdataJSON, fmt.Sprintf(JOB_USAGE, UID, info.JobID, "cpu"), "dimensions", "user", "value")
		info.CPU.Core["util"] = int64(cpuSysUtil) + int64(cpuUserUtil)
	}

	//TODO: modify to list later? (multi-node single job)
	// If that's the case, should multi netdataJSON, and cpu/mem util need to
	// summation over allocated nodes
	info.Node = append(info.Node, C.GoString(sData.nodes))
	gresCnt := int(sData.gres_detail_cnt)
	gresDetailArr := *(*[]*C.char)(unsafe.Pointer(&reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(sData.gres_detail_str)),
		Len:  gresCnt,
		Cap:  gresCnt,
	}))

	//TODO: We don't know the mapping between allocated nodes and gres_detail,
	//but since gres_cnt at most 1, so we hard code here
	for _, nd := range info.Node {
		if occupyNode, ok := nodeAccessMap[nd]; ok {
			gpus := []int{}
			if gresCnt == 1 {
				gpus = parseJobGres(C.GoString(gresDetailArr[0]))
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
				idx := []interface{}{nd, gpu}
				info.GPU.Idxs = append(info.GPU.Idxs, idx)
				//info.GPU.Idxs = append(info.GPU.Idxs, gpu)
				info.GPU.Models = append(info.GPU.Models, occupyNode.GPUs[gpu].Model)
				info.GPU.Util += int64(occupyNode.GPUs[gpu].Util)
				info.GPU.MEM["total"] += occupyNode.GPUs[gpu].MemTotal
				info.GPU.MEM["util"] += int64(occupyNode.GPUs[gpu].Mem)
			}
		}
	}
}
