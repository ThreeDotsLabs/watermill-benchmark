provider "google" {
  project     = "${var.project}"
  region      = "europe-north1"
}

resource "google_compute_instance" "default" {
  name         = "benchmark-${formatdate("YYYYMMDDhhmmss", timestamp())}"
  machine_type = "n1-highcpu-16"
  zone         = "europe-north1-a"

  boot_disk {
    initialize_params {
      size  = 128
      type = "pd-ssd"
      image = "ubuntu-2204-jammy-v20240927"
    }
  }

  network_interface {
    network = "default"
    access_config {}
  }

  metadata = {
    ssh-keys = "benchmark:${file(var.pub_key_path)}"
  }

  metadata_startup_script = "${file("startup.sh")}"
}

output "public_ip" {
  value = "${google_compute_instance.default.network_interface.0.access_config.0.nat_ip}"
}

variable "project" {}
variable "pub_key_path" {}
