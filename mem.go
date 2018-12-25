package main

type MEMtype struct {
	Total int64 `json:"total"`
	Alloc int64 `json:"alloc"`
	Util  int64 `json:"util"`
}
