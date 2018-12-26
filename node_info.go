package main

/*
#cgo CFLAGS: -g -Wall
#cgo LDFLAGS: -lslurm

#include <slurm/slurm.h>
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
	Hostname string    `json:"hostname"`
	Online   bool      `json:"online"`
	Normal   bool      `json:"normal"`
	GPUcnt   int       `json:"-"`
	GPUmodel string    `json:"-"`
	CPU      CPUtype   `json:"cpu"`
	GPUs     []GPUtype `json:"gpus"`
	MEM      MEMtype   `json:"mem"`
}

type NodeJobType struct {
	JobID string    `json:"jobid"`
	User  string    `json:"user"`
	GPUs  []GPUtype `json:"gpus"`
	Spot  bool      `json:"spot"`
}

func (info *NodeInfoType) setGres(gres string) { //gpu:1080Ti:2
	gres = gres[4:]
	info_ls := strings.Split(gres, ":")
	info.GPUmodel = info_ls[0]
	info.GPUcnt, _ = strconv.Atoi(info_ls[1])
}

func (info *NodeInfoType) init(netdataJSON []uint8, sData _Ctype_struct_node_info) {

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
		//TODO: need to modify later
		info.GPUs[i].MemTotal = 12300
		//info.Mem.Util = jsonparser.GetInt(netdataJSON, "system.ram", "dimensions", "used", "value")
		info.GPUs[i].Util, _ = jsonparser.GetFloat(netdataJSON, fmt.Sprintf("nvidia_smi.gpu%d_mem_utilization", i), "dimensions", fmt.Sprintf("gpu%d_memory_util", i), "value")
		info.GPUs[i].Mem, _ = jsonparser.GetFloat(netdataJSON, fmt.Sprintf("nvidia_smi.gpu%d_mem_allocated", i), "dimensions", fmt.Sprintf("gpu%d_fb_memory_usage", i), "value")
		info.GPUs[i].Temp, _ = jsonparser.GetFloat(netdataJSON, fmt.Sprintf("nvidia_smi.gpu%d_temperature", i), "dimensions", fmt.Sprintf("gpu%d_gpu_temp", i), "value")
	}
}
