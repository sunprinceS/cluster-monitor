package main

//import "C"
import (
	"fmt"
	"github.com/buger/jsonparser"
)

type NodeInfo struct {
	Hostname string    `json:"hostname"`
	Online   bool      `json:"online"`
	Normal   bool      `json:"normal"`
	GPUcnt   int       `json:"-"`
	GPUmodel string    `json:"-"`
	CPU      CPUtype   `json:"cpu"`
	GPUs     []GPUtype `json:"gpus"`
	MEM      MEMtype   `json:"mem"`
}

func (info *NodeInfo) init(netdataJSON []uint8) {
	// Set CPU
	info.CPU.Total = 16
	info.CPU.Alloc = 12
	cpuUtilTmp, _ := jsonparser.GetInt(netdataJSON, "system.cpu", "dimension", "user", "value")
	info.CPU.Util = float64(cpuUtilTmp*info.CPU.Total) / 100
	info.CPU.Temp, _ = jsonparser.GetInt(netdataJSON, "sensors.coretemp-isa-0000_temperature", "dimensions", "coretemp-isa-0000_temp1", "value")

	// Set Memory
	info.MEM.Total = 22200 - 11100
	info.MEM.Alloc = 222
	info.MEM.Util, _ = jsonparser.GetInt(netdataJSON, "system.ram", "dimensions", "used", "value")

	info.GPUcnt = 2
	info.GPUmodel = "2080Ti"
	info.GPUs = make([]GPUtype, info.GPUcnt)

	//Set GPU
	for i := 0; i < info.GPUcnt; i++ {
		info.GPUs[i].Model = info.GPUmodel
		//TODO: need to modify later
		info.GPUs[i].MemTotal = 12300
		//info.Mem.Util = jsonparser.GetInt(netdataJSON, "system.ram", "dimensions", "used", "value")
		info.GPUs[i].Util, _ = jsonparser.GetInt(netdataJSON, fmt.Sprintf("nvidia_smi.gpu%d_mem_utilization", i), "dimensions", fmt.Sprintf("gpu%d_memory_util", i), "value")
		info.GPUs[i].Mem, _ = jsonparser.GetInt(netdataJSON, fmt.Sprintf("nvidia_smi.gpu%d_mem_allocated", i), "dimensions", fmt.Sprintf("gpu%d_fb_memory_usage", i), "value")
		info.GPUs[i].Temp, _ = jsonparser.GetInt(netdataJSON, fmt.Sprintf("nvidia_smi.gpu%d_temperature", i), "dimensions", fmt.Sprintf("gpu%d_gpu_temp", i), "value")
	}
}
