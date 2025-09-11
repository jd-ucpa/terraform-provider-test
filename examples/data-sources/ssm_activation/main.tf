data "test_ssm_activation" "activation" {
  activation_id = "07dea21a-e2af-4915-a029-b1706045595a"
}

output "activation_details" {
  value = {
    id                  = data.test_ssm_activation.activation.id
    iam_role           = data.test_ssm_activation.activation.iam_role
    registration_limit = data.test_ssm_activation.activation.registration_limit
    registrations_count = data.test_ssm_activation.activation.registrations_count
    expiration_date    = data.test_ssm_activation.activation.expiration_date
    expired            = data.test_ssm_activation.activation.expired
    created_date       = data.test_ssm_activation.activation.created_date
  }
}
