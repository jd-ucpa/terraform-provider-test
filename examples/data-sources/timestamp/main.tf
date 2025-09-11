# Basic configuration - returns current UTC timestamp
data "test_timestamp" "current" {}

# Configuration with all parameters
data "test_timestamp" "full_params" {
  time_zone = "Europe/Paris"
  time_add {
    days    = 1
    hours   = 2
    minutes = 30
    seconds = 45
  }
}

# Outputs to display the results
output "current_timestamp" {
  description = "Current UTC timestamp"
  value       = data.test_timestamp.current.result
}

output "full_params_timestamp" {
  description = "Timestamp with timezone and time addition"
  value       = data.test_timestamp.full_params.result
}
