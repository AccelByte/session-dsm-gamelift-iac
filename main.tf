data "aws_caller_identity" "current" {}

module "eventbridge" {
  source         = "./modules/eventbridge"
  rule_name      = "gamelift-placement-events-rule"
  sqs_queue_arn  = module.sqs.sqs_queue_arn
  event_bus_name = "default"
}

module "iam" {
  source         = "./modules/iam"
  aws_account_id = data.aws_caller_identity.current.account_id
}

module "lambda" {
  source = "./modules/lambda"

  lambda_name       = "gamelift-event-processor"
  lambda_build_path = "${path.root}/lambda/build"
  handler           = "bootstrap"
  runtime           = "provided.al2023"
  sqs_queue_arn     = module.sqs.sqs_queue_arn
}

module "sqs" {
  source                     = "./modules/sqs"
  queue_name                 = "gamelift-event-queue"
  lambda_arn                 = module.lambda.lambda_arn
  message_retention_seconds  = 120
  visibility_timeout_seconds = 30
}

module "ssm" {
  source = "./modules/ssm"
}
