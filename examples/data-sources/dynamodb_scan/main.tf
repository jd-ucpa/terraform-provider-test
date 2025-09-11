# Basic DynamoDB scan example
data "test_dynamodb_scan" "simple_scan" {
  table_name = "test-table"
}

output "simple_scan_items" {
  value = data.test_dynamodb_scan.simple_scan.items
}

output "simple_scan_count" {
  value = data.test_dynamodb_scan.simple_scan.items_count
}

# Example with index and projection
data "test_dynamodb_scan" "index_scan" {
  table_name = "test-table"
  index_name = "test-index"
  projection_expression = "product_uuid, customer_uuid"
}

output "index_scan_items" {
  value = data.test_dynamodb_scan.index_scan.items
}

# Example with attribute aliases (without # prefix)
data "test_dynamodb_scan" "alias_scan" {
  table_name = "test-table"
  index_name = "test-index"
  projection_expression = "cu, pu"
  expression_attribute_names = {
    cu = "customer_uuid"
    pu = "product_uuid"
  }
}

output "alias_scan_items" {
  value = data.test_dynamodb_scan.alias_scan.items
}

# Example with filter expression and attribute values
data "test_dynamodb_scan" "filter_scan" {
  table_name = "test-table"
  index_name = "test-index"
  projection_expression = "cu, pu"
  filter_expression = "st = accepted"
  expression_attribute_names = {
    cu = "customer_uuid"
    pu = "product_uuid"
    st = "status"
  }
  expression_attribute_values = {
    accepted = "ACCEPTED"
  }
}

output "filter_scan_items" {
  value = data.test_dynamodb_scan.filter_scan.items
}
