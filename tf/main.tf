provider "google" {
  credentials = file(var.gcp_credentials_file)
  project     = var.gcp_project
  region      = var.gcp_region
}

resource "google_compute_instance" "vm_instance" {
  name         = "docker-vm"
  machine_type = "e2-medium"
  zone         = var.gcp_zone

  boot_disk {
    initialize_params {
      # Ubuntu 22.04 LTS
      image = "ubuntu-os-cloud/ubuntu-2204-lts"
    }
  }

  network_interface {
    network = "default"
    access_config {
      // Ephemeral IP
    }
  }

  metadata_startup_script = <<-EOT
    #!/bin/bash
    # Update and Install Docker
    apt-get update
    apt-get install -y docker.io git

    # Install Docker Compose
    curl -L "https://github.com/docker/compose/releases/download/v2.0.0/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
    chmod +x /usr/local/bin/docker-compose

    # Clone your repository
    git clone github.com/louisloechel/CloudServiceBenchmarking /home/ubuntu/CloudServiceBenchmarking

    # Assuming docker-compose.yml is at the root of your repository
    # Navigate to the repository directory
    cd /home/ubuntu/CloudServiceBenchmarking

    # Build and run docker-compose
    docker-compose build
    docker-compose up -d
  EOT

  service_account {
    scopes = ["cloud-platform"]
  }

  tags = ["http-server", "https-server"]
}

resource "google_compute_firewall" "firewall" {
  name    = "allow-http-https"
  network = "default"

  allow {
    protocol = "tcp"
    ports    = ["80", "443"]
  }

  source_ranges = ["0.0.0.0/0"]
}
