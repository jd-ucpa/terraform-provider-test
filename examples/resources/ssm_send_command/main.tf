resource "test_ssm_send_command" "example" {
  document_name = "AWS-RunShellScript"
  instance_ids  = ["i-1234567890abcdef0"]
  
  parameters = {
    "commands" = "echo 'Hello from Terraform!' && pwd && date"
  }
  
  comment = "Basic SSM command example"
}
