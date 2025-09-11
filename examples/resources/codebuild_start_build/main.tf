# Provider configuration
provider "test" {
  region = "eu-west-1"
  
  # Optional: use a specific AWS profile
  # profile = "my-aws-profile"
  
  # Optional: assume an IAM role
  # assume_role {
  #   role_arn = "arn:aws:iam::123456789012:role/MyRole"
  #   session_name = "terraform-session"
  # }
}

# Basic example of starting a CodeBuild build
resource "test_codebuild_start_build" "basic_example" {
  project_name = "my-codebuild-project"
}

# Example with environment variables
resource "test_codebuild_start_build" "with_env_vars" {
  project_name = "my-codebuild-project"
  
  environment_variables {
    name  = "BUILD_ENV"
    value = "production"
    type  = "PLAINTEXT"
  }
  
  environment_variables {
    name  = "VERSION"
    value = "1.0.0"
    # default type: PLAINTEXT
  }
  
  environment_variables {
    name  = "SECRET_VALUE"
    value = "arn:aws:secretsmanager:eu-west-1:123456789012:secret:my-secret"
    type  = "SECRETS_MANAGER"
  }
}

# Complete example with all parameters
resource "test_codebuild_start_build" "complete_example" {
  project_name = "my-codebuild-project"
  
  environment_variables {
    name  = "BUILD_ENV"
    value = "production"
    type  = "PLAINTEXT"
  }
  
  environment_variables {
    name  = "VERSION"
    value = "2.0.0"
  }
  
  environment_variables {
    name  = "PARAMETER_VALUE"
    value = "/myapp/database/host"
    type  = "PARAMETER_STORE"
  }
}

# Outputs to retrieve build information
output "basic_build_id" {
  description = "ID of the basic build"
  value       = test_codebuild_start_build.basic_example.build_id
}

output "basic_build_arn" {
  description = "ARN of the basic build"
  value       = test_codebuild_start_build.basic_example.build_arn
}

output "basic_build_number" {
  description = "Number of the basic build"
  value       = test_codebuild_start_build.basic_example.build_number
}

output "complete_build_info" {
  description = "Complete build information"
  value = {
    id           = test_codebuild_start_build.complete_example.build_id
    arn          = test_codebuild_start_build.complete_example.build_arn
    build_number = test_codebuild_start_build.complete_example.build_number
    project_name = test_codebuild_start_build.complete_example.build_project_name
  }
}
