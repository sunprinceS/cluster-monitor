package main

type GPUtype struct {
	Model    string  `json:"model"`
	Used     bool    `json:"used"`
	MemTotal int64   `json:"mem_total"`
	Util     float64 `json:"util"`
	Mem      float64 `json:"mem"`
	Temp     float64 `json:"temp"`
}
