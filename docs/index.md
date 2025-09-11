---
page_title: "Provider: terraform-provider-test"
description: |-
  The Test provider is used to interact with AWS services through SSM Send Command and STS Caller Identity. The provider needs to be configured with the proper credentials before it can be used.
---

# Test Provider

Adds helpful resources for AWS.

## Example Usage

```terraform
terraform {
  required_providers {
    test = {
      source  = "jd-ucpa/test"
    }
  }
}

# Example with profile
provider "test" {
  alias   = "with_profile"
  region  = "eu-west-1"
  profile = "xxxx"
}

# Example with assume_role
provider "test" {
  alias  = "with_assume_role"
  region = "eu-west-1"
  assume_role {
    role_arn = "arn:aws:iam::xxxx:role/yyyy"
  }
}
```
