package workflow

import (
	"errors"
	"math"
	"time"

	"github.com/saeed903/temporal_microservices"
	"github.com/saeed903/temporal_microservices/domain/square"
	"github.com/saeed903/temporal_microservices/domain/volume"
	deepcopy "github.com/ulule/deepcopier"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

type Parallelepiped struct {
	ID         string
	Length     float64
	Width      float64
	Heigth     float32
	BaseSquare float64
	Volume     float64
}

type CalculateParallelepipedWorkflowRequest struct {
	BatchSize       int
	Parallelepipeds []Parallelepiped
}

type CalculateParallelepipedWorkflowResponse struct {
	Parallelepipeds []Parallelepiped
}

func CalculateParallelepipedWorkflow(ctx workflow.Context, req CalculateParallelepipedWorkflowRequest) (resp CalculateParallelepipedWorkflowResponse, err error) {
	if len(req.Parallelepipeds) <= 0 {
		err = errors.New("there are no figures to process")
		return
	}

	if isIDsRepeat(req.Parallelepipeds) {
		err = errors.New("the ids of the figures in the input should be unique")
		return
	}

	if req.BatchSize <= 0 {
		err = errors.New("makeBatch size cannot be less or equal zero")
		return
	}

	batchCount := batchCount(len(req.Parallelepipeds), req.BatchSize)

	selector := workflow.NewNamedSelector(ctx, "select-parallelepiped-batches")
	var errOnBatch error
	cancelCtx, cancelHandler := workflow.WithCancel(ctx)

	count := 0
	squareMap := make(map[string]float64)
	volumeMap := make(map[string]float64)
	for i := 0; i < batchCount; i++ {
		batch := makeBatch(req.Parallelepipeds, i, req.BatchSize)

		future := processSquareAsync(cancelCtx, batch)
		selector.AddFuture(future, func(f workflow.Future) {
			respSquare := square.CalculateRectangleSquareResponse{}
			if err := f.Get(cancelCtx, &respSquare); err != nil {
				cancelHandler()
				errOnBatch = err
			} else {
				copyResult(squareMap, respSquare.Squares)
			}
		})
		count++

		future = processVolumeSync(cancelCtx, batch)
		selector.AddFuture(future, func(f workflow.Future) {
			respVolume := volume.CalculateParallelepipeVolumeResponse{}
			if err := f.Get(cancelCtx, &respVolume); err != nil {
				cancelHandler()
				errOnBatch = err
			} else {
				copyResult(volumeMap, respVolume.Volumes)
			}
		})
		count++
	}

	// wait until everything processd
	for i := 0; i < count; i++ {
		selector.Select(ctx)
		if errOnBatch != nil {
			return CalculateParallelepipedWorkflowResponse{}, errOnBatch
		}
	}

	// map the output
	var outputFigures = make([]Parallelepiped, 0, len(req.Parallelepipeds))
	for _, p := range req.Parallelepipeds {
		outputP := p
		outputP.BaseSquare = squareMap[p.ID]
		outputP.Volume = volumeMap[p.ID]
		outputFigures = append(outputFigures, outputP)
	}
	return CalculateParallelepipedWorkflowResponse{Parallelepipeds: outputFigures}, nil
}

func batchCount(wholeSize, batchSize int) int {
	if batchSize <= 0 {
		return 0
	}
	return int(math.Ceil(float64(wholeSize) / float64(batchSize)))
}

func copyResult(output map[string]float64, input map[string]float64) {
	for k, v := range input {
		output[k] = v
	}
}

func makeBatch(source []Parallelepiped, batchNumber, batchSize int) []Parallelepiped {
	start := batchNumber + batchSize
	end := start + batchSize
	atAll := len(source)
	if batchSize <= 0 || batchNumber < 0 || atAll < start {
		return []Parallelepiped{}
	}
	if end > atAll {
		end = atAll
	}
	return source[start:end]
}

func isIDsRepeat(Parallelepipeds []Parallelepiped) bool {
	idsMap := make(map[string]struct{})
	for _, p := range Parallelepipeds {
		_, ok := idsMap[p.ID]
		if ok {
			return true
		}
		idsMap[p.ID] = struct{}{}
	}
	return false
}

func processSquareAsync(cancelCtx workflow.Context, batch []Parallelepiped) workflow.Future {
	future, settable := workflow.NewFuture(cancelCtx)
	workflow.Go(cancelCtx, func(ctx workflow.Context) {
		ctx = withActivityOptions(ctx, temporal_microservices.SquareAcitivityQueue)
		respSquare := square.CalculateRectangleSquareResponse{}
		dimensions, err := copySquareBatch(batch)
		if err != nil {
			settable.Set(nil, err)
			return
		}
		err = workflow.ExecuteActivity(ctx,
			square.RectangleSquareActivityName,
			square.CalculateRectangleSquareRequest{Rectangles: dimensions}).Get(ctx, &respSquare)
		settable.Set(respSquare, err)
	})

	return future
}

func processVolumeSync(cancelCtx workflow.Context, batch []Parallelepiped) workflow.Future {
	future, settable := workflow.NewFuture(cancelCtx)
	workflow.Go(cancelCtx, func(ctx workflow.Context) {
		ctx = withActivityOptions(ctx, temporal_microservices.VolumeActivityQueue)
		respVolume := volume.CalculateParallelepipeVolumeResponse{}
		dimensions, err := copyVolumeBatch(batch)
		if err != nil {
			settable.Set(nil, err)
			return
		}
		err = workflow.ExecuteActivity(ctx,
			volume.ParallelepipeVolumeActivityName,
			volume.CalculateParallelepipeVolumeRequest{Parallelepipes: dimensions}).Get(ctx, &respVolume)
		settable.Set(respVolume, err)
	})
	return future
}

func copyVolumeBatch(source []Parallelepiped) (dest []volume.Parallelepipe, err error) {
	dest = make([]volume.Parallelepipe, len(source))
	for i, v := range source {
		if err = deepcopy.Copy(&v).To(&dest[i]); err != nil {
			return
		}
	}
	return
}
func copySquareBatch(source []Parallelepiped) (dest []square.Rectangle, err error) {
	dest = make([]square.Rectangle, len(source))
	for i, p := range source {
		if err = deepcopy.Copy(&p).To(&dest[i]); err != nil {
			return
		}
	}
	return
}

func withActivityOptions(ctx workflow.Context, queue string) workflow.Context {
	ao := workflow.ActivityOptions{
		TaskQueue:              queue,
		ScheduleToStartTimeout: 24 * time.Hour,
		StartToCloseTimeout:    24 * time.Hour,
		HeartbeatTimeout:       time.Second * 5,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:        time.Second,
			BackoffCoefficient:     2.0,
			MaximumInterval:        time.Minute * 5,
			NonRetryableErrorTypes: []string{"BusinessError"},
		},
	}

	ctxOut := workflow.WithActivityOptions(ctx, ao)
	return ctxOut

}
