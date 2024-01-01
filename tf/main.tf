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

  metadata_startup_script = file("startup.sh")

  service_account {
    scopes = ["cloud-platform"]
  }

  tags = ["http-server", "https-server"]

  metadata = {
    ssh-keys = "user_name:${file("../env/my-ssh-key.pub")}"
  }
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
      host        = google_compute_instance.vm_instance.network_interface[0].access_config[0].nat_ip
    }
  }

  provisioner "local-exec" {
    // Caution: "StrictHostKeyChecking=no"" is less secure but should be fine in this use case.
    command = "echo 'Benchmarking completed. Downloading results.csv.' && scp -o StrictHostKeyChecking=no -i ../env/my-ssh-key user_name@${google_compute_instance.vm_instance.network_interface[0].access_config[0].nat_ip}:/home/ubuntu/CloudServiceBenchmarking/results.csv ../results.csv"
  }
}
