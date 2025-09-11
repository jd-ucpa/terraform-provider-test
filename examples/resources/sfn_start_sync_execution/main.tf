# Provider configuration
provider "test" {
  region = "eu-west-1"
}

# Basic synchronous execution of an SFN state machine
resource "test_sfn_start_sync_execution" "basic" {
  state_machine_arn = "arn:aws:states:eu-west-1:123456789012:stateMachine:MyStateMachine"
}

# Synchronous execution with a custom name
resource "test_sfn_start_sync_execution" "with_name" {
  state_machine_arn = "arn:aws:states:eu-west-1:123456789012:stateMachine:MyStateMachine"
  name              = "my-custom-execution"
}

# Synchronous execution with custom JSON input
resource "test_sfn_start_sync_execution" "with_input" {
  state_machine_arn = "arn:aws:states:eu-west-1:123456789012:stateMachine:MyStateMachine"
  name              = "execution-with-data"
  input             = jsonencode({
    message = "Hello from Terraform"
    value   = 42
    data = {
      nested = "value"
    }
  })
}

# Synchronous execution with custom input
resource "test_sfn_start_sync_execution" "with_custom_input" {
  state_machine_arn = "arn:aws:states:eu-west-1:123456789012:stateMachine:MyStateMachine"
  name              = "execution-with-custom-input"
  input             = jsonencode({
    message = "Hello from Terraform"
    value   = 42
  })
}

# Synchronous execution with all attributes
resource "test_sfn_start_sync_execution" "complete" {
  state_machine_arn = "arn:aws:states:eu-west-1:123456789012:stateMachine:MyStateMachine"
  name              = "complete-execution"
  input             = jsonencode({
    user_id    = "user123"
    operation  = "process_data"
    parameters = {
      timeout = 300
      retries = 3
    }
  })
  
}

# Display results
output "basic_execution_arn" {
  description = "ARN of the basic execution"
  value       = test_sfn_start_sync_execution.basic.execution_arn
}

output "basic_execution_status" {
  description = "Status of the basic execution"
  value       = test_sfn_start_sync_execution.basic.status
}

output "basic_execution_output" {
  description = "Output of the basic execution"
  value       = test_sfn_start_sync_execution.basic.output
}

output "with_name_execution_arn" {
  description = "ARN of the execution with custom name"
  value       = test_sfn_start_sync_execution.with_name.execution_arn
}

output "with_input_execution_arn" {
  description = "ARN of the execution with custom input"
  value       = test_sfn_start_sync_execution.with_input.execution_arn
}

output "complete_execution_arn" {
  description = "ARN of the complete execution"
  value       = test_sfn_start_sync_execution.complete.execution_arn
}

output "complete_execution_billing" {
  description = "Billing details of the complete execution"
  value       = test_sfn_start_sync_execution.complete.billing_details
}

output "with_custom_input_execution_arn" {
  description = "ARN of the execution with custom input"
  value       = test_sfn_start_sync_execution.with_custom_input.execution_arn
}

# Example outputs for error fields (will be null on success)
output "basic_execution_error" {
  description = "Error of the basic execution (null on success)"
  value       = test_sfn_start_sync_execution.basic.error
}

output "basic_execution_cause" {
  description = "Cause of the basic execution error (null on success)"
  value       = test_sfn_start_sync_execution.basic.cause
}

output "complete_execution_error" {
  description = "Error of the complete execution (null on success)"
  value       = test_sfn_start_sync_execution.complete.error
}

output "complete_execution_cause" {
  description = "Cause of the complete execution error (null on success)"
  value       = test_sfn_start_sync_execution.complete.cause
}
