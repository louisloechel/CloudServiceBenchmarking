variable "gcp_credentials_file" {
  description = "Path to the GCP credentials JSON file."
  default     = "../env/csb-middleware-bench-5a30c5375d23.json"
}

variable "gcp_project" {
  description = "The GCP project ID."
  default     = "csb-middleware-bench"
}

variable "gcp_region" {
  description = "The GCP region."
  default     = "europe-west10"
}

variable "gcp_zone" {
  description = "The GCP zone."
  default     = "europe-west10-a"
}

variable "docker_images" {
  description = "Docker images to deploy."
  type        = list(string)
  default     = ["grpc-server", "grpc-client"]
}

variable "docker_compose_path" {
  description = "Path to the docker-compose.yml file on local machine."
  default     = "../docker-compose.yml"
}
