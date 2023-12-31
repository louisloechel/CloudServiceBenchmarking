package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/timeout"
	pb "github.com/louisloechel/cloudservicebenchmarking/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	address               = "localhost:50051"
	defaultName           = "world"
	totalRequests         = 100 // Total number of requests to send
	maxConcurrentRequests = 10  // Maximum number of concurrent requests
	minConcurrentRequests = 1   // Minimum number of concurrent requests
)

type Metric struct {
	Duration time.Duration
}

func runBenchmark(c pb.GreeterClient, concurrentRequests int) {
	log.Printf("Running benchmark with %d concurrent requests", concurrentRequests)

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

	// Open results.csv for appending, create it if it doesn't exist
	file, err := os.OpenFile("/results/results.csv", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Could not open results.csv: %v", err)
	}
	defer file.Close()

	// Check if the file is empty
	info, err := file.Stat()
	if err != nil {
		log.Fatalf("Could not get file info: %v", err)
	}

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header if the file is empty
	if info.Size() == 0 {
		err = writer.Write([]string{"Total Requests", "Concurrent Requests", "Average Latency", "Max Latency", "Min Latency"})
		if err != nil {
			log.Fatalf("Could not write to results.csv: %v", err)
		}
	}

	// Write data
	err = writer.Write([]string{
		fmt.Sprintf("%d", totalRequests),
		fmt.Sprintf("%d", concurrentRequests),
		fmt.Sprintf("%v", avgDuration),
		fmt.Sprintf("%v", maxDuration),
		fmt.Sprintf("%v", minDuration),
	})
	if err != nil {
		log.Fatalf("Could not write to results.csv: %v", err)
	}
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

	for concurrentRequests := minConcurrentRequests; concurrentRequests <= maxConcurrentRequests; concurrentRequests++ {
		runBenchmark(c, concurrentRequests)
	}
	// var wg sync.WaitGroup
	// metricsChan := make(chan Metric, totalRequests)
	// semaphore := make(chan struct{}, concurrentRequests)

	// for i := 0; i < totalRequests; i++ {
	// 	wg.Add(1)
	// 	semaphore <- struct{}{} // Blocks if concurrentRequests are already running
	// 	go func() {
	// 		defer wg.Done()
	// 		defer func() { <-semaphore }() // Release the semaphore

	// 		start := time.Now()
	// 		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	// 		defer cancel()
	// 		_, err := c.SayHello(ctx, &pb.HelloRequest{Name: defaultName})
	// 		if err != nil {
	// 			log.Printf("Could not greet: %v", err)
	// 			return
	// 		}
	// 		metricsChan <- Metric{Duration: time.Since(start)}
	// 	}()
	// }

	// wg.Wait()
	// close(metricsChan)

	// // Calculate and print metrics
	// var totalDuration time.Duration
	// var maxDuration time.Duration
	// var minDuration = time.Duration(1<<63 - 1)
	// count := 0

	// for metric := range metricsChan {
	// 	totalDuration += metric.Duration
	// 	if metric.Duration > maxDuration {
	// 		maxDuration = metric.Duration
	// 	}
	// 	if metric.Duration < minDuration {
	// 		minDuration = metric.Duration
	// 	}
	// 	count++
	// }

	// avgDuration := totalDuration / time.Duration(count)

	// log.Printf("Total requests: %d", totalRequests)
	// log.Printf("Concurrent requests: %d", concurrentRequests)
	// log.Printf("Average latency: %v", avgDuration)
	// log.Printf("Max latency: %v", maxDuration)
	// log.Printf("Min latency: %v", minDuration)

	// // Open results.csv for appending, create it if it doesn't exist
	// file, err := os.OpenFile("/results/results.csv", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	// if err != nil {
	// 	log.Fatalf("Could not open results.csv: %v", err)
	// }
	// defer file.Close()

	// // Check if the file is empty
	// info, err := file.Stat()
	// if err != nil {
	// 	log.Fatalf("Could not get file info: %v", err)
	// }

	// writer := csv.NewWriter(file)
	// defer writer.Flush()

	// // Write header if the file is empty
	// if info.Size() == 0 {
	// 	err = writer.Write([]string{"Total Requests", "Concurrent Requests", "Average Latency", "Max Latency", "Min Latency"})
	// 	if err != nil {
	// 		log.Fatalf("Could not write to results.csv: %v", err)
	// 	}
	// }

	// // Write data
	// err = writer.Write([]string{
	// 	fmt.Sprintf("%d", totalRequests),
	// 	fmt.Sprintf("%d", concurrentRequests),
	// 	fmt.Sprintf("%v", avgDuration),
	// 	fmt.Sprintf("%v", maxDuration),
	// 	fmt.Sprintf("%v", minDuration),
	// })
	// if err != nil {
	// 	log.Fatalf("Could not write to results.csv: %v", err)
	// }

}
