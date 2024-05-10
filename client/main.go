/*
DISCLAIMER: Parts of this code are referencing following sources:
- https://github.com/grpc-ecosystem/go-grpc-middleware/tree/main/examples
*/

package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/go-yaml/yaml"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/timeout"
	pb "github.com/louisloechel/cloudservicebenchmarking/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	grpcMetadata "google.golang.org/grpc/metadata"

	jwt "github.com/Siar-Akbayin/jwt-go-auth"
)

type Metric struct {
	Duration time.Duration
}

type Config struct {
	Address               string `yaml:"server_address"`
	Port                  string `yaml:"server_port"`
	DefaultName           string `yaml:"default_name"`
	TotalRequests         int    `yaml:"total_requests"`
	MaxConcurrentRequests int    `yaml:"max_concurrent_requests"`
	MinConcurrentRequests int    `yaml:"min_concurrent_requests"`
	WarmupRequests        int    `yaml:"warmup_requests"`
}

func loadConfig() Config {
	configFile := "./config.yml"
	configData, err := os.ReadFile(configFile)
	if err != nil {
		log.Fatalf("failed to read config file: %v", err)
	}
	log.Printf("Loaded config file: %v", configFile)
	log.Printf("Config data: %v", string(configData))

	// Parse the YAML data into the Config struct
	var config Config
	err = yaml.Unmarshal(configData, &config)
	if err != nil {
		log.Fatalf("failed to parse config file: %v", err)
	}

	// print config
	log.Printf("Config: %v", config)

	return config
}

func initialiseResultsFile() {
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
		err = writer.Write([]string{"Timestamp", "Total Requests", "Concurrent Requests", "Average Latency", "Max Latency", "Min Latency", "Avg. Throughput req/s", "Time Elapsed"})
		if err != nil {
			log.Fatalf("Could not write to results.csv: %v", err)
		}
	}

	// Write data
	err = writer.Write([]string{
		fmt.Sprintf("%v", time.Now()),
		fmt.Sprintf("%d", 0),
		fmt.Sprintf("%d", 0),
		fmt.Sprintf("%v", 0),
		fmt.Sprintf("%v", 0),
		fmt.Sprintf("%v", 0),
		fmt.Sprintf("%v", 0),
		fmt.Sprintf("%v", 0),
	})
	if err != nil {
		log.Fatalf("Could not write to results.csv: %v", err)
	}
}

func warmUp(c pb.GreeterClient, concurrentRequests int, config Config) {
	var wg sync.WaitGroup
	metricsChan := make(chan Metric, config.WarmupRequests)
	semaphore := make(chan struct{}, concurrentRequests)

	// generate token
	// GenerateToken(policyPath string, serviceName string, purpose string, keyPath string, expirationInHours time.Duration)
	token, err := jwt.GenerateToken("policy.json", "client", "purpose1", "private_key.pem", 1)
	// log.Printf("Token: %s", token)
	fmt.Println(token)
	if err != nil {
		log.Fatalf("Error on generating token: %v", err)
	}

	for i := 0; i < config.WarmupRequests; i++ {
		wg.Add(1)
		semaphore <- struct{}{} // Blocks if concurrentRequests are already running
		go func() {
			defer wg.Done()
			defer func() { <-semaphore }() // Release the semaphore

			// Custom auth
			md := grpcMetadata.Pairs("authorization", token)

			start := time.Now()
			ctx, cancel := context.WithTimeout(grpcMetadata.NewOutgoingContext(context.Background(), md), time.Second)
			// Append Metadata w/ Good Client Token
			ctx = grpcMetadata.AppendToOutgoingContext(ctx, "authorization", token)

			defer cancel()
			id := i%1000 + 1
			response, err := c.SayHello(ctx, &pb.HelloRequest{Name: config.DefaultName, Id: int32(id)})
			if err != nil {
				log.Printf("Could not greet: %v", err)
				return
			} else {
				log.Printf("Response: %v", response)
			}
			metricsChan <- Metric{Duration: time.Since(start)}
		}()
	}
}

func runBenchmark(c pb.GreeterClient, concurrentRequests int, config Config) {
	log.Printf("\nRunning benchmark with %d concurrent requests", concurrentRequests)
	runStart := time.Now()

	var wg sync.WaitGroup
	metricsChan := make(chan Metric, config.TotalRequests)
	semaphore := make(chan struct{}, concurrentRequests)

	// generate token
	// GenerateToken(policyPath string, serviceName string, purpose string, keyPath string, expirationInHours time.Duration)
	goodToken, err := jwt.GenerateToken("policy.json", "client", "purpose1", "private_key.pem", 1)
	if err != nil {
		log.Fatalf("Error on generating token: %v", err)
	}

	mixedToken, err := jwt.GenerateToken("policy.json", "client", "purpose2", "private_key.pem", 1)
	if err != nil {
		log.Fatalf("Error on generating token: %v", err)
	}

	badToken, err := jwt.GenerateToken("policy.json", "client", "purpose3", "private_key.pem", 1)
	if err != nil {
		log.Fatalf("Error on generating token: %v", err)
	}

	// three different mixed-tokens used to compare each anononymization technique's performance
	generalizedToken, err := jwt.GenerateToken("policy.json", "client", "purpose4", "private_key.pem", 1)
	if err != nil {
		log.Fatalf("Error on generating token: %v", err)
	}

	noisedToken, err := jwt.GenerateToken("policy.json", "client", "purpose5", "private_key.pem", 1)
	if err != nil {
		log.Fatalf("Error on generating token: %v", err)
	}

	reducedToken, err := jwt.GenerateToken("policy.json", "client", "purpose6", "private_key.pem", 1)
	if err != nil {
		log.Fatalf("Error on generating token: %v", err)
	}

	// log.Printf("Token: %s", token)
	// fmt.Println(token)
	if err != nil {
		log.Fatalf("Error on generating token: %v", err)
	}

	for i := 0; i < config.TotalRequests; i++ {
		wg.Add(1)
		semaphore <- struct{}{} // Blocks if concurrentRequests are already running
		go func() {
			defer wg.Done()
			defer func() { <-semaphore }() // Release the semaphore

			// Randomly select token
			var token string
			switch rand.Intn(6) {
			case 0:
				token = goodToken
			case 1:
				token = mixedToken
			case 2:
				token = badToken
			case 3:
				token = generalizedToken
			case 4:
				token = noisedToken
			case 5:
				token = reducedToken
			}

			// Deactivate random token selection
			// Options: generalizedToken, noisedToken, reducedToken, mixedToken, goodToken, badToken (see above)
			token = generalizedToken

			// Custom auth
			md := grpcMetadata.Pairs("authorization", token)

			start := time.Now()
			ctx, cancel := context.WithTimeout(grpcMetadata.NewOutgoingContext(context.Background(), md), time.Second)
			// Append Metadata w/ Good Client Token
			ctx = grpcMetadata.AppendToOutgoingContext(ctx, "authorization", token)

			defer cancel()
			id := i%1000 + 1 // id is between 1 and 1000 to not run out of the .csv file on the server side
			response, err := c.SayHello(ctx, &pb.HelloRequest{Name: config.DefaultName, Id: int32(id)})
			if err != nil {
				log.Printf("Could not greet: %v", err)
				return
			} else {
				log.Printf("Response: %v", response)
			}
			metricsChan <- Metric{Duration: time.Since(start)}
		}()
	}

	wg.Wait()
	close(metricsChan)

	// Calculate and print metrics.
	totalDuration := time.Since(runStart)
	var maxDuration time.Duration
	var minDuration = time.Duration(1<<63 - 1)
	var sumDuration time.Duration

	count := 0

	for metric := range metricsChan {
		if metric.Duration > maxDuration {
			maxDuration = metric.Duration
		}
		if metric.Duration < minDuration {
			minDuration = metric.Duration
		}
		count++
		sumDuration += metric.Duration
	}

	avgThroughput := float64(config.TotalRequests) / float64(sumDuration.Seconds()) //float64(totalDuration.Seconds())
	avgLatency := sumDuration / time.Duration(count)                                // totalDuration/time.Duration(config.TotalRequests)

	log.Printf("Total requests: %d", config.TotalRequests)
	log.Printf("Concurrent requests: %d", concurrentRequests)
	log.Printf("Average latency: %v", avgLatency)
	log.Printf("Max latency: %v", maxDuration)
	log.Printf("Min latency: %v", minDuration)
	log.Printf("Avg Throughput: %f req/s", avgThroughput)
	log.Printf("Time elapsed: %v", totalDuration)

	// // Moving average throughput
	// for _, throughput := range movingAvgThroughput {
	// 	log.Printf("Moving average throughput: %f req/s", throughput)
	// }

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
		err = writer.Write([]string{"Timestamp", "Total Requests", "Concurrent Requests", "Average Latency", "Max Latency", "Min Latency", "Avg. Throughput req/s", "Time Elapsed"})
		if err != nil {
			log.Fatalf("Could not write to results.csv: %v", err)
		}
	}

	// Write data
	err = writer.Write([]string{
		fmt.Sprintf("%v", time.Now()),
		fmt.Sprintf("%d", config.TotalRequests),
		fmt.Sprintf("%d", concurrentRequests),
		fmt.Sprintf("%v", avgLatency),
		fmt.Sprintf("%v", maxDuration),
		fmt.Sprintf("%v", minDuration),
		fmt.Sprintf("%v", avgThroughput),
		fmt.Sprintf("%v", totalDuration),
	})
	if err != nil {
		log.Fatalf("Could not write to results.csv: %v", err)
	}
}

func experimentDone() {
	// Create experiment_done.txt
	file, err := os.Create("/results/experiment_done.txt")
	if err != nil {
		log.Fatalf("Could not create experiment_done.txt: %v", err)
	}
	defer file.Close()
}

func main() {
	config := loadConfig()

	endpoint := fmt.Sprintf("%s:%s", config.Address, config.Port)

	conn, err := grpc.Dial(
		endpoint,
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

	// Warm up the server
	log.Printf("Warming up the server. Sending %d requests", config.WarmupRequests)
	warmUp(c, config.MaxConcurrentRequests, config)
	log.Printf("Warm up finished. Benchmarking...\n------------------")

	// initialise results.csv with start time and zeroes
	initialiseResultsFile()

	// Run the benchmark
	for concurrentRequests := config.MinConcurrentRequests; concurrentRequests <= config.MaxConcurrentRequests; concurrentRequests++ {
		runBenchmark(c, concurrentRequests, config)
	}

	// create indicator that benchmark is finished: experiment_done.txt
	experimentDone()
	log.Printf("Benchmark finished. Created experiment_done.txt")
}
