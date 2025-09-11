# Example usage of the json_pretty data source

# Basic usage - format JSON with default parameters
data "test_json_pretty" "basic" {
  json = jsonencode({
    name = "John"
    age  = 30
    city = "New York"
  })
}

# Format JSON with all parameters
data "test_json_pretty" "full_params" {
  json = jsonencode({
    config = {
      database = {
        host = "localhost"
        port = 5432
      }
      cache = {
        enabled = true
        ttl     = 3600
      }
    }
  })
  indent  = 4
  newline = true
}

# Outputs to display the formatted results
output "basic_json_pretty" {
  description = "Basic JSON formatting with default parameters"
  value       = data.test_json_pretty.basic.result
}

output "full_params_json_pretty" {
  description = "JSON formatting with all parameters (indent=4, newline=true)"
  value       = data.test_json_pretty.full_params.result
}
