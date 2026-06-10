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
  # client_id     = "..."   # or set CIRRO_CLIENT_ID
  # client_secret = "..."   # or set CIRRO_CLIENT_SECRET (recommended)
}
