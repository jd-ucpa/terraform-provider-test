# Basic usage - get all metric alarms
data "test_cloudwatch_alarms" "basic" {
  alarm_type = "MetricAlarm"
}

# Filter by alarm name prefix
data "test_cloudwatch_alarms" "alarm_name_prefix" {
  alarm_name_prefix = "test-"
}

# Get specific alarms by name
data "test_cloudwatch_alarms" "alarm_names" {
  alarm_names = ["test-alarm-1", "test-alarm-2"]
}

# Filter by state value
data "test_cloudwatch_alarms" "state_value" {
  state_value = "INSUFFICIENT_DATA"
}

# Get composite alarms
data "test_cloudwatch_alarms" "composite" {
  alarm_type = "CompositeAlarm"
}

# Include Auto Scaling alarms (default behavior is to ignore them)
data "test_cloudwatch_alarms" "include_autoscaling" {
  ignore_autoscaling_alarms = false
}

# Combined filters
data "test_cloudwatch_alarms" "combined" {
  alarm_type = "MetricAlarm"
  alarm_name_prefix = "prod-"
  state_value = "ALARM"
  ignore_autoscaling_alarms = true
}
