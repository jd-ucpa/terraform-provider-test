terraform {
  required_providers {
    test = {
      source  = "jd-ucpa/test"
      version = "0.0.3"
    }
  }
}

# Configuration with specific timezone
data "test_timestamp" "paris_time" {
  time_zone = "Europe/Paris"
}

# Outputs to display the results
output "current_timestamp" {
  description = "Current UTC timestamp"
  value       = data.test_timestamp.paris_time
}
