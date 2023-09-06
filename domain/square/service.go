package square

import (
	"context"

	"github.com/saeed903/temporal_microservices"
	"github.com/saeed903/temporal_microservices/domain"
)

var RectangleSquareActivityName = domain.GetActivityName(Service{}.CalculateRectangleSquare)

type Rectangle struct {
	ID     string
	Length float64
	Width  float64
}

type CalculateRectangleSquareRequest struct {
	Rectangles []Rectangle
}

type CalculateRectangleSquareResponse struct {
	Squares map[string]float64
}

type Service struct{}

func (s Service) CalculateRectangleSquare(ctx context.Context, req CalculateRectangleSquareRequest) (resp CalculateRectangleSquareResponse, err error) {
	hearbeat := domain.StartHeartbeat(ctx, temporal_microservices.HeartbeatIntervalSec)
	defer hearbeat.Stop()

	resp.Squares = make(map[string]float64, len(req.Rectangles))
	for _, r := range req.Rectangles {
		resp.Squares[r.ID] = r.Width * r.Length
	}
	return
}
