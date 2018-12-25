package main

type GPUtype struct {
	Model    string `json:"model"`
	Used     bool   `json:"used"`
	MemTotal int64  `json:"mem_total"`
	Util     int64  `json:"util"`
	Mem      int64  `json:"mem"`
	Temp     int64  `json:"temp"`
}
