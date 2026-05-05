# Terraform Provider for Productive.io

A Terraform/OpenTofu provider for managing resources in [Productive.io](https://productive.io).

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0 or [OpenTofu](https://opentofu.org/) >= 1.9
- [Go](https://golang.org/doc/install) >= 1.24 (for building from source)

## Authentication

The provider requires a Productive.io API token and organization ID. These can be provided via:

1. Provider configuration block
2. Environment variables: `PRODUCTIVE_TOKEN` and `PRODUCTIVE_ORGANIZATION_ID`

Generate an API token at **Settings > API Integrations > Generate new token** in Productive.io.

## Usage

```hcl
terraform {
  required_providers {
    productive = {
      source  = "productive-io/productive"
      version = "~> 0.1.0"
    }
  }
}

provider "productive" {
  # Credentials can also be set via PRODUCTIVE_TOKEN and PRODUCTIVE_ORGANIZATION_ID env vars.
  token           = var.productive_token
  organization_id = var.productive_organization_id
}

resource "productive_person" "example" {
  first_name = "Jane"
  last_name  = "Smith"
  email      = "jane.smith@example.com"
  title      = "Engineer"
  role_id    = 3
}
```

## Resources

- `productive_person` — Manages a person (user) in Productive.io

## Importing Existing Resources

```shell
terraform import productive_person.example 1042276
```

> **Note:** Destroying a `productive_person` resource archives the person in Productive.io (soft delete). Productive.io does not support hard deletion via the API.

## Building from Source

```shell
go install
```

## Development

```shell
make build       # compile
make test        # unit tests
make testacc     # acceptance tests (requires PRODUCTIVE_TOKEN and PRODUCTIVE_ORGANIZATION_ID)
make lint        # golangci-lint
make generate    # regenerate docs
```

To use a locally built provider, add a dev override to `~/.terraformrc` or `~/.tofurc`:

```hcl
provider_installation {
  filesystem_mirror {
    path = "/path/to/go/tf-plugins"
  }
  direct {}
}
```

## License

MPL-2.0
