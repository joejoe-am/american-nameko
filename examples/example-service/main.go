package main

import (
	"context"
	"fmt"
	"github.com/joejoe-am/namego/configs"
	"github.com/joejoe-am/namego/examples/example-service/service"
	"github.com/joejoe-am/namego/pkg/rpc"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

// TODO: change package name

func main() {
	cfg := configs.GetConfigs()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	var wg sync.WaitGroup

	amqpConnection := InitRabbitMQ(cfg.RabbitMQURL)
	defer amqpConnection.Close()

	// RPC client example
	err := rpc.InitClient(amqpConnection)

	if err != nil {
		log.Fatal(err)
	}

	authRpc := rpc.NewClient("authnzng")
	quotaRpc := rpc.NewClient("quota")

	if err != nil {
		log.Fatalf("failed to initialize RPC clients: %v", err)
	}

	response, err := authRpc.CallRpc("health_check", map[string]string{})
	fmt.Println(response, err)

	response, err = quotaRpc.CallRpc("health_check", map[string]string{})
	fmt.Println(response, err)

	// RPC server example
	rpcServer := rpc.NewServer("nameko", amqpConnection)
	rpcServer.RegisterMethod("multiply", service.Multiply)

	//handlerConfig := events.EventConfig{
	//	SourceService:    "authnzng",
	//	EventType:        "EVENT_EXAMPLE",
	//	HandlerType:      events.ServicePool,
	//	ReliableDelivery: true,
	//	HandlerFunction:  service.EventHandlerFunction,
	//}
	//
	//eventHandler, err := events.NewEventHandler(handlerConfig)
	//if err != nil {
	//	log.Fatalf("failed to create event handler: %v", err)
	//}
	//
	//err = eventHandler.Start(amqpConnection)
	//if err != nil {
	//	log.Fatalf("failed to start event handler: %v", err)
	//}

	// Dispatch event Example
	service.DispatchEventExampleFunction(amqpConnection, cfg.ServiceName)

	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Println("starting rpc server")
		if err := rpcServer.Start(); err != nil {
			log.Printf("RPC server error: %v", err)
			cancel()
		}
	}()

	select {
	case sig := <-signalChan:
		log.Printf("Received signal: %v", sig)
		cancel()
	case <-ctx.Done():
	}

	//shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	//defer shutdownCancel()
	//if err := server.Shutdown(shutdownCtx); err != nil {
	//	log.Printf("Error shutting down web server: %v", err)
	//}
	//
	//if err := rpcServer.Stop(); err != nil {
	//	log.Printf("Error shutting down RPC server: %v", err)
	//}
	//
	//if err := eventHandler.Close(); err != nil {
	//	log.Printf("Error shutting down event handler: %v", err)
	//}
	//
	//wg.Wait()

	log.Println("Shutdown complete.")
}
