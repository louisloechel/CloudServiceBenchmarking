# CloudServiceBenchmarking
**🧪 Benchmarking gRPC interceptors for Microservices**

## Rodamap
- [ ] Set-up simple client-server architecture (containerized)
- [ ] add variety of interceptors
- [ ] create load generator/benchmarking tool
- [ ] GCP setup
- [ ] run Benchmarking Experiments
- [ ] Analyze results
- [ ] Write report

## Local Setup
### Compile Locally
From the ```/server``` directory of this repo, compile the server binary.
```bash
go build
```



### Build Docker Images
From the top of this repo, build the docker images for server and client.

```bash
docker build -t grpc-server -f server/Dockerfile .
```
```bash
docker build -t grpc-client -f client/Dockerfile .
```
### Run containers
Run the server container first, and then run the client container. The client container needs to be run with the `--network="host"` flag to allow it to connect to the server container.
``` bash
docker run -p 50051:50051 grpc-server
```    
``` bash
docker run --network="host" grpc-client
```

...

## GCP Setup
_coming soon_