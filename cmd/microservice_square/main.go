package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/saeed903/temporal_microservices"
	"github.com/saeed903/temporal_microservices/domain/square"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

func main() {
	log.Print("starting SQUARE microservice...")

	temporalClient := initTemporalClient()

	worker := initActivityWorker(temporalClient)

	signals := make(chan os.Signal, 1)

	signal.Notify(signals, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	<-signals

	worker.Stop()

	log.Print("closing SQUARE microservice.")

}

func initTemporalClient() client.Client {
	temporalClientOptions := client.Options{
		HostPort: net.JoinHostPort("localhost", "7233")}
	temporalClient, err := client.NewClient(temporalClientOptions)
	if err != nil {
		log.Fatal("cannot start temporal client: " + err.Error())
	}
	return temporalClient
}

func initActivityWorker(temporalClient client.Client) worker.Worker {
	workerOption := worker.Options{
		MaxConcurrentActivityExecutionSize: temporal_microservices.MaxConcurrentSquareActivitySize,
	}
	worker := worker.New(temporalClient, temporal_microservices.SquareAcitivityQueue, workerOption)
	worker.RegisterActivity(square.Service{}.CalculateRectangleSquare)

	err := worker.Start()
	if err != nil {
		log.Fatal("cannot start temporal worker: " + err.Error())
	}
	return worker
}
