output "gamelift_aws_access_key" {
  value     = module.iam.gamelift_aws_access_key
  sensitive = true
}

output "gamelift_aws_secret_key" {
  value     = module.iam.gamelift_aws_secret_key
  sensitive = true
}
