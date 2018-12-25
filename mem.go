package main

type MEMtype struct {
	Total int64   `json:"total"`
	Alloc int64   `json:"alloc"`
	Util  float64 `json:"util"`
}
