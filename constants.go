package temporal_microservices

const (
	SquareAcitivityQueue = "SquareActivityQueue"
	VolumeActivityQueue  = "VolumeActivityQueue"
	FigureWorkflowQueue  = "FigureWorkflowQueue"

	MaxConcurrentSquareActivitySize = 10
	MaxConcurrentVolumeActivitySize = 10
	MaxConcurrentFigureWorkflowSize = 3

	HeartbeatIntervalSec = 1
)
