# CloudServiceBenchmarking
**🧪 Benchmarking gRPC interceptors for Microservices**

## Rodamap
- [X] Set-up simple client-server architecture (containerized)
- [ ] add variety of interceptors
- [ ] create load generator/benchmarking tool
- [X] GCP setup
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