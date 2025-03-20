output "session_dsm_access_key" {
  value     = module.iam.session_dsm_access_key
  sensitive = true
}

output "session_dsm_secret_key" {
  value     = module.iam.session_dsm_secret_key
  sensitive = true
}
