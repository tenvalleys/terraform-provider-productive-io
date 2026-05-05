provider "productive" {
  # Credentials can also be set via PRODUCTIVE_TOKEN and PRODUCTIVE_ORGANIZATION_ID env vars.
  token           = var.productive_token
  organization_id = var.productive_organization_id
}
