# CloudServiceBenchmarking
**🧪 Benchmarking gRPC Interceptors for Microservices**

## Rodamap
- [X] Set-up simple client-server architecture (containerized)
- [X] add variety of interceptors
- [X] create load generator/benchmarking tool
- [X] GCP setup
- [ ] run Benchmarking Experiments
- [ ] Analyze results
- [ ] Write report

## Benchmark Configuration
### Server
Specify which interceptor to include on the server side by toggling the respective boolean in the ```server/config.yml``` file (for full list, check out the file). 

Example for the Prometheus Metrics Interceptor:
```yml
# Prometheus Metrics Interceptor
metrics:
  enabled: true
```

### Client
Define the load pattern to generate in the ```client/config.yml``` file. In the example below, the client will generate a load of 100.000 requests over 100 iterations each, resulting in 10.000.000 requests throughout the whole experiment. With each iteration the number of concurrent (parallel) requests is increased.
```yml
total_requests: 100000
max_concurrent_requests: 100
min_concurrent_requests: 1
```
#### Warmup
Specify the amount of requests sentto the server before the experiment is prperly started in the ```client/config.yml``` file. In the example below, the client will send 100.000 requests before the experiment is started.
```yml
warmup_requests: 1000000
```

## Local Setup
### Compile Locally
From the ```/server``` directory of this repo, compile the server binary.
```bash
go build
```



### Build Docker Images
From the top of this repo, build the docker images for server and client.

```bash
docker-compose build
```
### Run containers
Runs the server container first, and then the client container. The client container needs to be run with the `--network="host"` flag to allow it to connect to the server container.
``` bash
docker-compose up
```
### Bring down containers
```bash
docker-compose down
```
...

## GCP Setup
### Terraform
From the ```/tf``` directory of this repo, initialise terraform
```bash
terraform init
```
Then run the terraform script to create the GCP resources
```bash
terraform apply
```
To delete the resources, run
```bash
terraform destroy
```