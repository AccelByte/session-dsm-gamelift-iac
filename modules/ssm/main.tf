resource "aws_ssm_parameter" "ab_base_url" {
  name        = "/lambda/ab_base_url"
  description = "AccelByte API Base URL"
  type        = "String"
  value       = "placeholder"

  lifecycle {
    ignore_changes = [value]
  }
}

resource "aws_ssm_parameter" "ab_client_id" {
  name        = "/lambda/ab_client_id"
  description = "AccelByte Client ID"
  type        = "String"
  value       = "placeholder"

  lifecycle {
    ignore_changes = [value]
  }
}

resource "aws_ssm_parameter" "ab_client_secret" {
  name        = "/lambda/ab_client_secret"
  description = "AccelByte Client Secret"
  type        = "String"
  value       = "placeholder"

  lifecycle {
    ignore_changes = [value]
  }
}

resource "aws_ssm_parameter" "ab_namespace_name" {
  name        = "/lambda/ab_namespace_name"
  description = "AccelByte Namespace Name"
  type        = "String"
  value       = "placeholder"

  lifecycle {
    ignore_changes = [value]
  }
}
