package main

/*
#cgo CFLAGS: -g -Wall
#cgo LDFLAGS: -lslurm

#include <slurm/slurm.h>
*/
import "C"
import (
	//"fmt"
	"strconv"
	"strings"
	"time"
)

//TODO: modify to the same as KaiChi's version?
//func formatTime(t time.Time){
//}

// Turn minutes to string
func time2str(mins int64) string {
	if mins == int64(C.INFINITE) {
		return "UNLIMITED"
	} else {
		return (time.Duration(mins) * time.Minute).String()
	}
}

func parseGresIdx(idxAllS string) (ret []int) {
	//e.g 0-2,3,5-7
	for _, idxS := range strings.Split(idxAllS, ",") {
		idxList := strings.Split(idxS, "-")
		if len(idxList) == 1 {
			idx, _ := strconv.Atoi(idxList[0])
			ret = append(ret, idx)
		} else {
			start, _ := strconv.Atoi(idxList[0])
			end, _ := strconv.Atoi(idxList[1])
			for idx := start; idx <= end; idx++ {
				ret = append(ret, idx)
			}
		}
	}
	return
}
