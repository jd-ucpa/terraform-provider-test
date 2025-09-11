resource "test_ssm_send_files" "example" {
  platform = "linux"
  instance_ids = ["i-1234567890abcdef0"]
  working_directory = "/tmp"
  
  script_before_files = "echo 'Starting file creation...' && pwd"
  script_after_files = "echo 'Files created successfully' && ls -la test_file.txt && cat test_file.txt"
  
  file {
    name = "test_file.txt"
    content = "Hello from Terraform SSM Send Files!"
    permissions = "644"
    owner = "ec2-user"
    group = "ec2-user"
  }
}
