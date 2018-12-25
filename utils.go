package main

import (
	//"fmt"
	"strconv"
	"strings"
)

// TODO: maybe move to node_info
func parseGresIdx(gres string) (ret []int) {
	gres = strings.Split(gres[:len(gres)-1], "IDX")[1][1:]
	if gres != "N/A" {
		//segStrList = strings.Split(gres,",")
		for _, segStr := range strings.Split(gres, ",") {
			segStrLs := strings.Split(segStr, "-")
			if len(segStrLs) == 1 {
				ele, _ := strconv.Atoi(segStrLs[0])
				ret = append(ret, ele)
			} else {
				start, _ := strconv.Atoi(segStrLs[0])
				end, _ := strconv.Atoi(segStrLs[1])
				for i := start; i <= end; i++ {
					ret = append(ret, i)
				}
			}
		}
	}
	return
}
