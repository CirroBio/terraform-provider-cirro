# Terraform Provider for Cirro

This Terraform provider allows you to manage [Cirro](https://cirro.bio) resources via infrastructure as code.

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.7
- [Go](https://golang.org/doc/install) >= 1.22 (to build from source)

## Using the Provider

```hcl
terraform {
  required_providers {
    cirro = {
      source  = "cirro-bio/cirro"
      version = "~> 0.1"
    }
  }
}

provider "cirro" {
  base_url = "https://app.cirro.bio"
}
```

### Authentication

The provider authenticates using OAuth2 client credentials. Set credentials via environment variables (recommended):

```sh
export CIRRO_BASE_URL="https://app.cirro.bio"
export CIRRO_CLIENT_ID="your-client-id"
export CIRRO_CLIENT_SECRET="your-client-secret"
```

Or configure inline (not recommended for secrets in production):

```hcl
provider "cirro" {
  base_url      = "https://app.cirro.bio"
  client_id     = "your-client-id"
  client_secret = "your-client-secret"
}
```

The provider automatically exchanges the client credentials for a Bearer token at
`{base_url}/auth/token` and refreshes it before expiry.

## Resources

| Resource | Description |
|---|---|
| `cirro_project` | Create and manage projects |
| `cirro_project_member` | Manage user roles within a project |
| `cirro_user` | Invite and manage users |

## Data Sources

| Data Source | Description |
|---|---|
| `data.cirro_project` | Look up a project by ID |
| `data.cirro_user` | Look up a user by username |

## Building from Source

```sh
git clone https://github.com/cirro-bio/terraform-provider-cirro
cd terraform-provider-cirro
go mod tidy
make install
```

## Developing

### Running Tests

```sh
make test       # unit tests
make testacc    # acceptance tests (requires CIRRO_BASE_URL, CIRRO_CLIENT_ID, CIRRO_CLIENT_SECRET)
```

### Generating Docs

```sh
make generate
```
