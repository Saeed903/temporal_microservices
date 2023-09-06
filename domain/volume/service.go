package volume

import (
	"context"

	"github.com/saeed903/temporal_microservices"
	"github.com/saeed903/temporal_microservices/domain"
)

var ParallelepipeVolumeActivityName = domain.GetActivityName(Service{}.CalculateParallelepipeVolume)

type Parallelepipe struct {
	ID     string
	Length float64
	Width  float64
	Heigth float64
}

type Service struct{}

type CalculateParallelepipeVolumeRequest struct {
	Parallelepipes []Parallelepipe
}

type CalculateParallelepipeVolumeResponse struct {
	Volumes map[string]float64
}

func (s Service) CalculateParallelepipeVolume(ctx context.Context, req CalculateParallelepipeVolumeRequest) (resp CalculateParallelepipeVolumeResponse, err error) {
	heartbeat := domain.StartHeartbeat(ctx, temporal_microservices.HeartbeatIntervalSec)
	defer heartbeat.Stop()

	resp.Volumes = make(map[string]float64, len(req.Parallelepipes))

	for _, p := range req.Parallelepipes {
		resp.Volumes[p.ID] = p.Heigth * p.Length * p.Width
	}
	return
}
