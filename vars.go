package main

const (
	API_VERSION      = 1
	PORT             = 19999
	DOMAIN           = "speech"
	DELAY            = 5 // seconds
	INSTANCE         = "http://localhost:%d/host/%s.%s/api/v%d"
	ALL_METRIC_POINT = "allmetrics?format=json&help=no&types=no&timestamps=yes&names=yes&data=average"
	JOB_USAGE        = "cgroup_slurm_uid_%s_job_%s.%s"
)
