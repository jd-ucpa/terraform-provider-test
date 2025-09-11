# Example usage of the test_ssm_activation resource

# Basic SSM activation
resource "test_ssm_activation" "basic" {
  iam_role    = "SSMServiceRole"
  secret_name = "my/activation/secret"
  description = "Basic SSM activation for hybrid instances"
}

# SSM activation with custom expiration date
resource "test_ssm_activation" "with_expiration" {
  iam_role    = "SSMServiceRole"
  secret_name = "my/activation/secret"
  description = "SSM activation with custom expiration"
  
  expiration_date {
    days    = 10
    hours   = 8
    minutes = 6
  }
}

# SSM activation with secret management
resource "test_ssm_activation" "with_secret" {
  iam_role    = "SSMServiceRole"
  description = "SSM activation with secret management"
  secret_name = "my/activation/code"
  managed     = true
}

# SSM activation with tags
resource "test_ssm_activation" "with_tags" {
  iam_role    = "SSMServiceRole"
  secret_name = "my/activation/secret"
  description = "SSM activation with tags"
  
  tags = {
    Environment = "production"
    Project     = "hybrid-cloud"
    Owner       = "platform-team"
  }
}

# SSM activation with registration limit
resource "test_ssm_activation" "with_limit" {
  iam_role          = "arn:aws:iam::123456789012:role/SSMServiceRole"
  secret_name       = "my/activation/limit"
  description       = "SSM activation with registration limit"
  registration_limit = 5
}

# Outputs
output "basic_activation_id" {
  description = "The ID of the basic SSM activation"
  value       = test_ssm_activation.basic.activation_id
}

output "basic_activation_code" {
  description = "The activation code of the basic SSM activation"
  value       = test_ssm_activation.basic.activation_code
  sensitive   = true
}

output "secret_activation_arn" {
  description = "The ARN of the secret containing activation data"
  value       = test_ssm_activation.with_secret.secret_arn
}
