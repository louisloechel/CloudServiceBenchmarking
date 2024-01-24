provider "google" {
  credentials = file(var.gcp_credentials_file)
  project     = var.gcp_project
  region      = var.gcp_region
}

# Create a VPC
resource "google_compute_network" "vpc_network" {
  name                    = "csb-vpc"
  auto_create_subnetworks = false
}

# Create a subnet
resource "google_compute_subnetwork" "subnet" {
  name          = "csb-subnet"
  network       = google_compute_network.vpc_network.name
  ip_cidr_range = "10.0.0.0/16"
  region        = var.gcp_region
}

# Create the server VM instance
resource "google_compute_instance" "server_instance" {
  name         = "server-vm"
  machine_type = "e2-standard-4" # 16 vCPUs, 64 GB memory | no shared resources and enough of 'em
  zone         = var.gcp_zone

  boot_disk {
    initialize_params {
      # Ubuntu 22.04 LTS
      image = "ubuntu-os-cloud/ubuntu-2204-lts"
    }
  }

  network_interface {
    subnetwork = google_compute_subnetwork.subnet.self_link
    access_config {
      // Ephemeral IP
    }
  }

  metadata_startup_script = file("startup_server.sh")

  service_account {
    scopes = ["cloud-platform"]
  }

  tags = ["http-client", "https-client"]

  metadata = {
    ssh-keys = "user_name:${file("../env/my-ssh-key.pub")}"
  }
}

# Create the client VM instance
resource "google_compute_instance" "client_instance" {
  name         = "client-vm"
  machine_type = "e2-standard-16" # 16 vCPUs, 64 GB memory | no shared resources and enough of 'em
  zone         = var.gcp_zone

  boot_disk {
    initialize_params {
      # Ubuntu 22.04 LTS
      image = "ubuntu-os-cloud/ubuntu-2204-lts"
    }
  }

  network_interface {
    subnetwork = google_compute_subnetwork.subnet.self_link
    access_config {
      // Ephemeral IP
    }
  }

  metadata_startup_script = file("startup_client.sh")

  service_account {
    scopes = ["cloud-platform"]
  }

  tags = ["http-server", "https-server"]

  metadata = {
    ssh-keys = "user_name:${file("../env/my-ssh-key.pub")}"
  }
}

# Create a firewall rule to allow internal communication within the VPC
resource "google_compute_firewall" "allow_internal" {
  name    = "allow-internal"
  network = google_compute_network.vpc_network.name

  allow {
    protocol = "tcp"
    ports    = ["0-65535"]
  }

  allow {
    protocol = "udp"
    ports    = ["0-65535"]
  }

  allow {
    protocol = "icmp"
  }

  source_ranges = ["10.0.0.0/16"]
}

# Create a firewall rule to allow external SSH, HTTP, HTTPS access
resource "google_compute_firewall" "allow_external" {
  name    = "allow-external"
  network = google_compute_network.vpc_network.name

  allow {
    protocol = "tcp"
    ports    = ["22", "80", "443"]
  }

  source_ranges = ["0.0.0.0/0"]
}

# Download result.csv from the client VM instance to the local machine after the benchmark is done
resource "null_resource" "benchmark_waiter" {
  triggers = {
    always_run = "${timestamp()}"
  }

  provisioner "remote-exec" {
    inline = [
      "while [ ! -f /home/ubuntu/CloudServiceBenchmarking/experiment_done.txt ]; do echo 'Conducting Benchmark...'; sleep 10; done",
    ]

    connection {
      type        = "ssh"
      user        = "user_name"
      private_key = file(var.private_key_path)
      host        = google_compute_instance.client_instance.network_interface[0].access_config[0].nat_ip
    }
  }

  provisioner "local-exec" {
    // >>!! CAUTION !!<< "StrictHostKeyChecking=no"" is less secure but should be fine in this use case.
    command = "echo 'Benchmarking completed. Downloading results.csv.' && scp -o StrictHostKeyChecking=no -i ../env/my-ssh-key user_name@${google_compute_instance.client_instance.network_interface[0].access_config[0].nat_ip}:/home/ubuntu/CloudServiceBenchmarking/results.csv ../results.csv"
  }
}
