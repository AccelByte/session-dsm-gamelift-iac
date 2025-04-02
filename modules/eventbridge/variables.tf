variable "rule_name" {
  description = "The name of the EventBridge rule"
  type        = string
  default     = "gamelift-placement-events-rule"
}

variable "event_bus_name" {
  description = "The name of the EventBridge bus"
  type        = string
  default     = "default"
}

variable "sqs_queue_arn" {
  description = "The ARN of the SQS queue to send filtered GameLift events"
  type        = string
}
