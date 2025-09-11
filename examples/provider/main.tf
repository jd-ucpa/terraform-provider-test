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
