package main

import "C"

type NodeInfo struct {
	hostname string
	online   bool
	normal   bool
	cpu      Cpu
	mem      Mem
	gpus     []Gpu
}
