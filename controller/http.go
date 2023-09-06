package controller

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/saeed903/temporal_microservices"
	"github.com/saeed903/temporal_microservices/domain/workflow"
	"go.temporal.io/sdk/client"
)

func MakeFigureHandleFunc(temporalClient client.Client) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()
		wr, err := getWorkflowRequest(r)
		if err != nil {
			writeError(w, err)
			return
		}
		output, err := executeWorkflow(ctx, temporalClient, wr)
		if err != nil {
			writeError(w, err)
			return
		}
		err = writeOutputFigures(w, output)
		if err != nil {
			writeError(w, err)
			return
		}

	}
}

func writeOutputFigures(w http.ResponseWriter, output []workflow.Parallelepiped) error {
	body, err := json.Marshal(output)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(body)
	if err != nil {
		return err
	}
	return nil

}

func executeWorkflow(ctx context.Context, temporalClient client.Client, wr workflow.CalculateParallelepipedWorkflowRequest) (output []workflow.Parallelepiped, err error) {
	workflowOptions := client.StartWorkflowOptions{
		TaskQueue: temporal_microservices.FigureWorkflowQueue,
	}

	workflowRun, err := temporalClient.ExecuteWorkflow(ctx, workflowOptions, workflow.CalculateParallelepipedWorkflow, wr)
	if err != nil {
		return
	}
	worflowResp := workflow.CalculateParallelepipedWorkflowResponse{}
	err = workflowRun.Get(ctx, &worflowResp)
	if err != nil {
		return
	}
	return worflowResp.Parallelepipeds, nil
}

func writeError(w http.ResponseWriter, err error) {
	log.Print(err.Error())
	w.WriteHeader(http.StatusInternalServerError)
	if _, errWrite := w.Write([]byte(err.Error())); errWrite != nil {
		log.Print("error writing the HTTP response: " + errWrite.Error())
	}
}

func getWorkflowRequest(r *http.Request) (wr workflow.CalculateParallelepipedWorkflowRequest, err error) {
	defer func() {
		if closeErr := r.Body.Close(); closeErr != nil {
			log.Print("error closing HTTP body: " + closeErr.Error())
		}
	}()
	if err = json.NewDecoder(r.Body).Decode(&wr); err != nil {
		log.Print("error decode request: " + err.Error())
	}
	return

}
