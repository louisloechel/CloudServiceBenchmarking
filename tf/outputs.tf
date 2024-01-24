output "client_public_ip" {
  value = google_compute_instance.client_instance.network_interface[0].access_config[0].nat_ip
}

output "server_public_ip" {
  value = google_compute_instance.server_instance.network_interface[0].access_config[0].nat_ip

}

output "server_internal_ip" {
  value = google_compute_instance.server_instance.network_interface[0].network_ip
}

