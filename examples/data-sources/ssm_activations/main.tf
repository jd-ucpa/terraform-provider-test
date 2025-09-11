# Get all non-expired activations
data "test_ssm_activations" "non_expired" {
  expired = false
}

# Get all expired activations
data "test_ssm_activations" "expired" {
  expired = true
}

# Get activations by specific IDs
data "test_ssm_activations" "by_ids" {
  filter {
    name = "activation-ids"
    values = ["07dea21a-e2af-4915-a029-b1706045595a"]
  }
}

# Get activations by IAM role
data "test_ssm_activations" "by_role" {
  filter {
    name = "iam-role"
    values = ["hybrid-activation"]
  }
}

# Combine filters (cumulative)
data "test_ssm_activations" "combined" {
  expired = true
  filter {
    name = "activation-ids"
    values = ["07dea21a-e2af-4915-a029-b1706045595a"]
  }
  filter {
    name = "iam-role"
    values = ["hybrid-activation"]
  }
}

output "activation_details" {
  value = data.test_ssm_activations.combined.activations
}
