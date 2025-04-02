variable "queue_name" {
  description = "The name of the SQS queue"
  type        = string
}

variable "message_retention_seconds" {
  description = "The number of seconds messages are kept in the queue"
  type        = number
  default     = 120 # Number of seconds before Accelbyte will mark the session failed on their end, invalidating the need for the message
}

variable "visibility_timeout_seconds" {
  description = "The duration in seconds that the message will be hidden from other consumers"
  type        = number
  default     = 30
}

variable "lambda_arn" {
  description = "The ARN of the Lambda function to be triggered"
  type        = string
}
