package main

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/timeout"
	pb "github.com/louisloechel/cloudservicebenchmarking/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	address            = "localhost:50051"
	defaultName        = "world"
	totalRequests      = 100000 // Total number of requests to send
	concurrentRequests = 1      // Number of concurrent requests
)

type Metric struct {
	Duration time.Duration
}

func main() {
	conn, err := grpc.Dial(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(
			timeout.UnaryClientInterceptor(500*time.Millisecond),
		),
	)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	c := pb.NewGreeterClient(conn)

	var wg sync.WaitGroup
	metricsChan := make(chan Metric, totalRequests)
	semaphore := make(chan struct{}, concurrentRequests)

	for i := 0; i < totalRequests; i++ {
		wg.Add(1)
		semaphore <- struct{}{} // Blocks if concurrentRequests are already running
		go func() {
			defer wg.Done()
			defer func() { <-semaphore }() // Release the semaphore

			start := time.Now()
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			_, err := c.SayHello(ctx, &pb.HelloRequest{Name: defaultName})
			if err != nil {
				log.Printf("Could not greet: %v", err)
				return
			}
			metricsChan <- Metric{Duration: time.Since(start)}
		}()
	}

	wg.Wait()
	close(metricsChan)

	// Calculate and print metrics
	var totalDuration time.Duration
	var maxDuration time.Duration
	var minDuration = time.Duration(1<<63 - 1)
	count := 0

	for metric := range metricsChan {
		totalDuration += metric.Duration
		if metric.Duration > maxDuration {
			maxDuration = metric.Duration
		}
		if metric.Duration < minDuration {
			minDuration = metric.Duration
		}
		count++
	}

	avgDuration := totalDuration / time.Duration(count)

	log.Printf("Total requests: %d", totalRequests)
	log.Printf("Concurrent requests: %d", concurrentRequests)
	log.Printf("Average latency: %v", avgDuration)
	log.Printf("Max latency: %v", maxDuration)
	log.Printf("Min latency: %v", minDuration)
}
