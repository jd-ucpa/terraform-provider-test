terraform {
  required_providers {
    test = {
      source = "jd-ucpa/test"
      version = "0.1.0"
    }
  }
}

provider "test" {
  # Ce provider ne nécessite aucune configuration
}

resource "test_do_nothing" "example" {
  name = "example-resource"
}
