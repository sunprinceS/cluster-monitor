package main

/*
#cgo CFLAGS: -g -Wall
#cgo LDFLAGS: -lslurm

#include <slurm/slurm.h>
// The following is exported from src/common/slurm_protocol_defs.h
char *node_state_string(uint32_t inx)
{
	int  base            = (inx & NODE_STATE_BASE);
	bool comp_flag       = (inx & NODE_STATE_COMPLETING);
	bool drain_flag      = (inx & NODE_STATE_DRAIN);
	bool fail_flag       = (inx & NODE_STATE_FAIL);
	bool maint_flag      = (inx & NODE_STATE_MAINT);
	bool net_flag        = (inx & NODE_STATE_NET);
	bool reboot_flag     = (inx & NODE_STATE_REBOOT);
	bool res_flag        = (inx & NODE_STATE_RES);
	bool resume_flag     = (inx & NODE_RESUME);
	bool no_resp_flag    = (inx & NODE_STATE_NO_RESPOND);
	bool power_down_flag = (inx & NODE_STATE_POWER_SAVE);
	bool power_up_flag   = (inx & NODE_STATE_POWER_UP);

	if (maint_flag) {
		if ((base == NODE_STATE_ALLOCATED) ||
		    (base == NODE_STATE_DOWN) ||
		    (base == NODE_STATE_MIXED))
			;
		else if (no_resp_flag)
			return "MAINT*";
		else
			return "MAINT";
	}
	if (reboot_flag) {
		if ((base == NODE_STATE_ALLOCATED) ||
		    (base == NODE_STATE_MIXED))
			;
		else if (no_resp_flag)
			return "REBOOT*";
		else
			return "REBOOT";
	}
	if (drain_flag) {
		if (comp_flag
		    || (base == NODE_STATE_ALLOCATED)
		    || (base == NODE_STATE_MIXED)) {
			if (maint_flag)
				return "DRAINING$";
			if (reboot_flag)
				return "DRAINING@";
			if (power_up_flag)
				return "DRAINING#";
			if (power_down_flag)
				return "DRAINING~";
			if (no_resp_flag)
				return "DRAINING*";
			return "DRAINING";
		} else {
			if (maint_flag)
				return "DRAINED$";
			if (reboot_flag)
				return "DRAINED@";
			if (power_up_flag)
				return "DRAINED#";
			if (power_down_flag)
				return "DRAINED~";
			if (no_resp_flag)
				return "DRAINED*";
			return "DRAINED";
		}
	}
	if (fail_flag) {
		if (comp_flag || (base == NODE_STATE_ALLOCATED)) {
			if (no_resp_flag)
				return "FAILING*";
			return "FAILING";
		} else {
			if (no_resp_flag)
				return "FAIL*";
			return "FAIL";
		}
	}

	if (inx == NODE_STATE_CANCEL_REBOOT)
		return "CANCEL_REBOOT";
	if (inx == NODE_STATE_POWER_SAVE)
		return "POWER_DOWN";
	if (inx == NODE_STATE_POWER_UP)
		return "POWER_UP";
	if (base == NODE_STATE_DOWN) {
		if (maint_flag)
			return "DOWN$";
		if (reboot_flag)
			return "DOWN@";
		if (power_up_flag)
			return "DOWN#";
		if (power_down_flag)
			return "DOWN~";
		if (no_resp_flag)
			return "DOWN*";
		return "DOWN";
	}

	if (base == NODE_STATE_ALLOCATED) {
		if (maint_flag)
			return "ALLOCATED$";
		if (reboot_flag)
			return "ALLOCATED@";
		if (power_up_flag)
			return "ALLOCATED#";
		if (power_down_flag)
			return "ALLOCATED~";
		if (no_resp_flag)
			return "ALLOCATED*";
		if (comp_flag)
			return "ALLOCATED+";
		return "ALLOCATED";
	}
	if (comp_flag) {
		if (maint_flag)
			return "COMPLETING$";
		if (reboot_flag)
			return "COMPLETING@";
		if (power_up_flag)
			return "COMPLETING#";
		if (power_down_flag)
			return "COMPLETING~";
		if (no_resp_flag)
			return "COMPLETING*";
		return "COMPLETING";
	}
	if (base == NODE_STATE_IDLE) {
		if (maint_flag)
			return "IDLE$";
		if (reboot_flag)
			return "IDLE@";
		if (power_up_flag)
			return "IDLE#";
		if (power_down_flag)
			return "IDLE~";
		if (no_resp_flag)
			return "IDLE*";
		if (net_flag)
			return "PERFCTRS";
		if (res_flag)
			return "RESERVED";
		return "IDLE";
	}
	if (base == NODE_STATE_MIXED) {
		if (maint_flag)
			return "MIXED$";
		if (reboot_flag)
			return "MIXED@";
		if (power_up_flag)
			return "MIXED#";
		if (power_down_flag)
			return "MIXED~";
		if (no_resp_flag)
			return "MIXED*";
		return "MIXED";
	}
	if (base == NODE_STATE_FUTURE) {
		if (maint_flag)
			return "FUTURE$";
		if (reboot_flag)
			return "FUTURE@";
		if (power_up_flag)
			return "FUTURE#";
		if (power_down_flag)
			return "FUTURE~";
		if (no_resp_flag)
			return "FUTURE*";
		return "FUTURE";
	}
	if (resume_flag)
		return "RESUME";
	if (base == NODE_STATE_UNKNOWN) {
		if (no_resp_flag)
			return "UNKNOWN*";
		return "UNKNOWN";
	}
	return "?";
}

extern char *node_state_string_compact(uint32_t inx)
{
	bool comp_flag       = (inx & NODE_STATE_COMPLETING);
	bool drain_flag      = (inx & NODE_STATE_DRAIN);
	bool fail_flag       = (inx & NODE_STATE_FAIL);
	bool maint_flag      = (inx & NODE_STATE_MAINT);
	bool net_flag        = (inx & NODE_STATE_NET);
	bool reboot_flag     = (inx & NODE_STATE_REBOOT);
	bool res_flag        = (inx & NODE_STATE_RES);
	bool resume_flag     = (inx & NODE_RESUME);
	bool no_resp_flag    = (inx & NODE_STATE_NO_RESPOND);
	bool power_down_flag = (inx & NODE_STATE_POWER_SAVE);
	bool power_up_flag   = (inx & NODE_STATE_POWER_UP);

	inx = (inx & NODE_STATE_BASE);

	if (maint_flag) {
		if ((inx == NODE_STATE_ALLOCATED) ||
		    (inx == NODE_STATE_DOWN) ||
		    (inx == NODE_STATE_MIXED))
			;
		else if (no_resp_flag)
			return "MAINT*";
		else
			return "MAINT";
	}
	if (reboot_flag) {
		if ((inx == NODE_STATE_ALLOCATED) || (inx == NODE_STATE_MIXED))
			;
		else if (no_resp_flag)
			return "BOOT*";
		else
			return "BOOT";
	}
	if (drain_flag) {
		if (comp_flag
		    || (inx == NODE_STATE_ALLOCATED)
		    || (inx == NODE_STATE_MIXED)) {
			if (maint_flag)
				return "DRNG$";
			if (reboot_flag)
				return "DRNG@";
			if (power_up_flag)
				return "DRNG#";
			if (power_down_flag)
				return "DRNG~";
			if (no_resp_flag)
				return "DRNG*";
			return "DRNG";
		} else {
			if (maint_flag)
				return "DRAIN$";
			if (reboot_flag)
				return "DRAIN@";
			if (power_up_flag)
				return "DRAIN#";
			if (power_down_flag)
				return "DRAIN~";
			if (no_resp_flag)
				return "DRAIN*";
			return "DRAIN";
		}
	}
	if (fail_flag) {
		if (comp_flag || (inx == NODE_STATE_ALLOCATED)) {
			if (no_resp_flag)
				return "FAILG*";
			return "FAILG";
		} else {
			if (no_resp_flag)
				return "FAIL*";
			return "FAIL";
		}
	}

	if (inx == NODE_STATE_CANCEL_REBOOT)
		return "CANC_R";
	if (inx == NODE_STATE_POWER_SAVE)
		return "POW_DN";
	if (inx == NODE_STATE_POWER_UP)
		return "POW_UP";
	if (inx == NODE_STATE_DOWN) {
		if (maint_flag)
			return "DOWN$";
		if (reboot_flag)
			return "DOWN@";
		if (power_up_flag)
			return "DOWN#";
		if (power_down_flag)
			return "DOWN~";
		if (no_resp_flag)
			return "DOWN*";
		return "DOWN";
	}

	if (inx == NODE_STATE_ALLOCATED) {
		if (maint_flag)
			return "ALLOC$";
		if (reboot_flag)
			return "ALLOC@";
		if (power_up_flag)
			return "ALLOC#";
		if (power_down_flag)
			return "ALLOC~";
		if (no_resp_flag)
			return "ALLOC*";
		if (comp_flag)
			return "ALLOC+";
		return "ALLOC";
	}
	if (comp_flag) {
		if (maint_flag)
			return "COMP$";
		if (reboot_flag)
			return "COMP@";
		if (power_up_flag)
			return "COMP#";
		if (power_down_flag)
			return "COMP~";
		if (no_resp_flag)
			return "COMP*";
		return "COMP";
	}
	if (inx == NODE_STATE_IDLE) {
		if (maint_flag)
			return "IDLE$";
		if (reboot_flag)
			return "IDLE@";
		if (power_up_flag)
			return "IDLE#";
		if (power_down_flag)
			return "IDLE~";
		if (no_resp_flag)
			return "IDLE*";
		if (net_flag)
			return "NPC";
		if (res_flag)
			return "RESV";
		return "IDLE";
	}
	if (inx == NODE_STATE_MIXED) {
		if (maint_flag)
			return "MIX$";
		if (reboot_flag)
			return "MIX@";
		if (power_up_flag)
			return "MIX#";
		if (power_down_flag)
			return "MIX~";
		if (no_resp_flag)
			return "MIX*";
		return "MIX";
	}
	if (inx == NODE_STATE_FUTURE) {
		if (maint_flag)
			return "FUTR$";
		if (reboot_flag)
			return "FUTR@";
		if (power_up_flag)
			return "FUTR#";
		if (power_down_flag)
			return "FUTR~";
		if (no_resp_flag)
			return "FUTR*";
		return "FUTR";
	}
	if (resume_flag)
		return "RESM";
	if (inx == NODE_STATE_UNKNOWN) {
		if (no_resp_flag)
			return "UNK*";
		return "UNK";
	}
	return "?";
}
*/
import "C"
import (
	"fmt"
	"github.com/buger/jsonparser"
	"strconv"
	"strings"
	"unsafe"
)

type NodeInfoType struct {
	Hostname string        `json:"hostname"`
	Online   bool          `json:"online"`
	Normal   bool          `json:"normal"`
	State    string        `json:"state"`
	GPUcnt   int           `json:"-"`
	GPUmodel string        `json:"-"`
	CPU      CPUtype       `json:"cpu"`
	GPUs     []GPUtype     `json:"gpus"`
	MEM      MEMtype       `json:"mem"`
	Jobs     []NodeJobType `json:"jobs"`
}

type NodeJobType struct {
	JobID string `json:"jobid"`
	User  string `json:"user"`
	GPUs  []int  `json:"gpus"`
	Spot  bool   `json:"spot"`
}

func (info *NodeInfoType) setGres(gres string) { //gpu:1080Ti:2
	gres = gres[4:]
	info_ls := strings.Split(gres, ":")
	info.GPUmodel = info_ls[0]
	info.GPUcnt, _ = strconv.Atoi(info_ls[1])
}

func checkNormal(state string) bool {
	abnormalLs := []string{"down", "drain", "drng", "fail", "failg"}
	for _, abState := range abnormalLs {
		if state == abState {
			return false
		}
	}
	return true
}
func (info *NodeInfoType) init(netdataJSON []uint8, sData _Ctype_struct_node_info) {

	info.State = strings.ToLower(C.GoString(C.node_state_string(sData.node_state)))
	info.Normal = checkNormal(info.State)
	info.setGres(C.GoString(sData.gres))
	gpuList := parseGresIdx(C.GoString(sData.gres_used))
	info.GPUs = make([]GPUtype, info.GPUcnt)
	if len(gpuList) > 0 {
		fmt.Println(gpuList)
		for _, idx := range gpuList {
			info.GPUs[idx].Used = true
		}
	}

	//gresUsedStr := C.GoString(sData.gres_used)
	info.Hostname = C.GoString(sData.node_hostname)

	// Set CPU
	info.CPU.Total = int64(sData.cpus)
	var cpuAllocTmp int64
	C.slurm_get_select_nodeinfo(sData.select_nodeinfo, C.SELECT_NODEDATA_SUBCNT, C.NODE_STATE_ALLOCATED, unsafe.Pointer(&cpuAllocTmp))
	info.CPU.Alloc = cpuAllocTmp
	//cpuUtilTmp, _ := jsonparser.GetInt(netdataJSON, "system.cpu", "dimension", "user", "value")
	//info.CPU.Util = float64(cpuUtilTmp*info.CPU.Total) / 100.0
	info.CPU.Util = float64(sData.cpu_load) / 100.0
	info.CPU.Temp, _ = jsonparser.GetFloat(netdataJSON, "sensors.coretemp-isa-0000_temperature", "dimensions", "coretemp-isa-0000_temp1", "value")

	// Set Memory
	info.MEM.Total = int64(sData.real_memory) - int64(sData.mem_spec_limit)
	var memAllocTmp int64
	C.slurm_get_select_nodeinfo(sData.select_nodeinfo, C.SELECT_NODEDATA_MEM_ALLOC, C.NODE_STATE_ALLOCATED, unsafe.Pointer(&memAllocTmp))
	info.MEM.Alloc = memAllocTmp
	info.MEM.Util, _ = jsonparser.GetFloat(netdataJSON, "system.ram", "dimensions", "used", "value")

	//Set GPU
	for i := 0; i < info.GPUcnt; i++ {
		info.GPUs[i].Model = info.GPUmodel
		//TODO: need to modify later (node config generation should handle)
		info.GPUs[i].MemTotal = 12300
		//info.Mem.Util = jsonparser.GetInt(netdataJSON, "system.ram", "dimensions", "used", "value")
		info.GPUs[i].Util, _ = jsonparser.GetFloat(netdataJSON, fmt.Sprintf("nvidia_smi.gpu%d_mem_utilization", i), "dimensions", fmt.Sprintf("gpu%d_memory_util", i), "value")
		info.GPUs[i].Mem, _ = jsonparser.GetFloat(netdataJSON, fmt.Sprintf("nvidia_smi.gpu%d_mem_allocated", i), "dimensions", fmt.Sprintf("gpu%d_fb_memory_usage", i), "value")
		info.GPUs[i].Temp, _ = jsonparser.GetFloat(netdataJSON, fmt.Sprintf("nvidia_smi.gpu%d_temperature", i), "dimensions", fmt.Sprintf("gpu%d_gpu_temp", i), "value")
	}
}
