output "gamelift_aws_access_key" {
  value     = aws_iam_access_key.session_dsm.id
  sensitive = true
}

output "gamelift_aws_secret_key" {
  value     = aws_iam_access_key.session_dsm.secret
  sensitive = true
}
