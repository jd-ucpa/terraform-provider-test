# Basic usage - get metrics by namespace and metric name
data "test_cloudwatch_metrics" "basic" {
  namespace = "AWS/EC2"
  metric_name = "CPUUtilization"
}

# Filter by namespace only
data "test_cloudwatch_metrics" "namespace_only" {
  namespace = "AWS/EC2"
}

# Filter by metric name only
data "test_cloudwatch_metrics" "metric_name_only" {
  metric_name = "CPUUtilization"
}

# Filter with dimensions
data "test_cloudwatch_metrics" "with_dimensions" {
  namespace = "AWS/EC2"
  metric_name = "CPUUtilization"
  
  dimension {
    name = "InstanceId"
    value = "i-0123456789abcdef0"
  }
}

# Filter with multiple dimensions
data "test_cloudwatch_metrics" "multiple_dimensions" {
  namespace = "AWS/EC2"
  
  dimension {
    name = "InstanceId"
    value = "i-0123456789abcdef0"
  }
  
  dimension {
    name = "AutoScalingGroupName"
    value = "test-asg"
  }
}

# Get all metrics (no filters)
data "test_cloudwatch_metrics" "all_metrics" {
}
