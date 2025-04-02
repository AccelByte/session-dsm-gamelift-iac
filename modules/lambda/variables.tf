variable "lambda_name" {
  description = "The name of the Lambda function"
  type        = string
}

variable "lambda_build_path" {
  description = "The path where the Lambda ZIP file is located"
  type        = string
}

variable "handler" {
  description = "The handler for the Lambda function"
  type        = string
  default     = "main"
}

variable "runtime" {
  description = "The runtime for the Lambda function"
  type        = string
  default     = "go1.x"
}

variable "sqs_queue_arn" {
  description = "The ARN of the SQS queue to trigger the Lambda function"
  type        = string
}
