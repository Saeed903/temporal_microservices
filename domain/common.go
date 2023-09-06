package domain

import (
	"context"
	"reflect"
	"runtime"
	"strings"
	"time"

	"go.temporal.io/sdk/activity"
)

func GetActivityName(i interface{}) string {

	fullName := runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
	parts := strings.Split(fullName, ".")
	lastPart := parts[len(parts)-1]
	parts2 := strings.Split(lastPart, "_")
	firstPart := parts2[0]
	return firstPart
}

type SimpleHearbit struct {
	cancelc chan struct{}
}

func StartHeartbeat(ctx context.Context, period time.Duration) (h *SimpleHearbit) {
	h = &SimpleHearbit{make(chan struct{})}
	go func() {
		for {
			select {
			case <-h.cancelc:
				return
			default:
				{
					time.Sleep(period)
					activity.RecordHeartbeat(ctx)
				}
			}
		}
	}()
	return h
}

func (h *SimpleHearbit) Stop() {
	close(h.cancelc)
}
