# Check Point - Workforce AI Terraform Provider
[![License](https://img.shields.io/github/license/CheckPointSW/terraform-provider-checkpoint-workforce-ai.svg?style=plastic)](https://github.com/CheckPointSW/terraform-provider-checkpoint-workforce-ai/blob/main/LICENSE) [![Tests](https://github.com/CheckPointSW/terraform-provider-checkpoint-workforce-ai/actions/workflows/test.yml/badge.svg)](https://github.com/CheckPointSW/terraform-provider-checkpoint-workforce-ai/actions/workflows/test.yml) [![Terraform Registry](https://img.shields.io/badge/terraform-registry-blueviolet.svg?style=plastic)](https://registry.terraform.io/providers/CheckPointSW/checkpoint-workforce-ai/latest)

This is the Terraform provider for managing **Check Point Workforce AI** and **Check Point Browse Security** policies as code.

The provider is based on the public [AI Security OpenAPI](https://app.swaggerhub.com/apis/Check-Point/checkpoint-ai-security) and [Browse Security OpenAPI](https://app.swaggerhub.com/apis/Check-Point/checkpoint-browse-security) specifications.

## Installation

Add the provider to your Terraform configuration:

```hcl
terraform {
  required_providers {
    cpwai = {
      source  = "CheckPointSW/checkpoint-workforce-ai"
      version = "~> 1.0"
    }
  }
}
```

Then run:

```bash
terraform init
```

## Getting started

Configure the provider with your CloudInfra API credentials and region.

```hcl
provider "cpwai" {
  client_id  = var.checkpoint_client_id
  access_key = var.checkpoint_access_key
  region     = "us"
}
```

### Obtaining API credentials

Depending on which product you want to manage, create the appropriate API key:

#### For Workforce AI

1. Go to the [Infinity Portal API Keys page](https://portal.checkpoint.com/dashboard/settings/api-keys).
2. Click **New** > **New Account API Key**.
3. In the **Service** dropdown select **Workforce AI Security** and create the key.
4. Copy the **Client ID** and **Secret Key** (Access Key).

#### For Browse Security

1. Go to the [Infinity Portal API Keys page](https://portal.checkpoint.com/dashboard/settings/api-keys).
2. Click **New** > **New Account API Key**.
3. In the **Service** dropdown select **Browser Security** and create the key.
4. Copy the **Client ID** and **Secret Key** (Access Key).

For more information, see [Infinity Portal Administration Guide](https://sc1.checkpoint.com/documents/Infinity_Portal/WebAdminGuides/EN/Infinity-Portal-Admin-Guide/Content/Topics-Infinity-Portal/API-Keys.htm?tocpath=Global%20Settings%7C_____7#API_Keys).

### Available gateways

| Region | Gateway URL |
|---|---|
| Europe (`eu`) | `https://cloudinfra-gw.portal.checkpoint.com` |
| United States (`us`) | `https://cloudinfra-gw-us.portal.checkpoint.com` |

## Complete example

```hcl
terraform {
  required_providers {
    cpwai = {
      source  = "CheckPointSW/checkpoint-workforce-ai"      version = "~> 1.0"
    }
  }
}

variable "checkpoint_client_id" {
  type      = string
  sensitive = true
}

variable "checkpoint_access_key" {
  type      = string
  sensitive = true
}

provider "cpwai" {
  client_id  = var.checkpoint_client_id
  access_key = var.checkpoint_access_key
  region     = "us"
}

# Block uploading credit card numbers to AI services
resource "cpwai_workforce_ai_chats_rule" "block_credit_cards" {
  name        = "Block Credit Card Uploads"
  description = "Prevent PII leakage to AI services"
  order       = 0
  active      = true

  policy = jsonencode({
    event_type = "file_upload"
    action     = "prevent"
    logging    = "enabled"
    services_and_application = {
      mode = "all"
    }
    data_types = [
      {
        id   = "cf0523c1-537e-4a4b-8bb8-084b7b9e0b45"
        name = "Credit Card Number"
        type = "PRE_DEFINED"
      }
    ]
  })

  source = [
    {
      assignment_type = "ASSIGNMENT_TYPE_ENTIRE_ORG"
    }
  ]
}
```

Set credentials via environment variables:

```bash
export TF_VAR_checkpoint_client_id="your-client-id"
export TF_VAR_checkpoint_access_key="your-access-key"
```

Then apply:

```bash
terraform init
terraform plan
terraform apply
```

## Resources

### AI Security

| Resource | Description |
|----------|-------------|
| `cpwai_workforce_ai_chats_rule` | DLP rules for AI chat services |
| `cpwai_workforce_ai_access_rule` | Access control rules for AI services |
| `cpwai_workforce_ai_agents_rule` | Policy rules for AI agents and MCP servers |

### Browse Security

| Resource | Description |
|----------|-------------|
| `cpwai_browse_dlp_rule` | DLP rules for web browsing services |
| `cpwai_browse_access_rule` | Access control rules for web browsing |
| `cpwai_browse_secure_browsing_rule` | Secure browsing rules (phishing, password reuse) |
| `cpwai_browse_file_protection_object` | Reusable file protection policy objects |
| `cpwai_browse_domains_object` | Reusable domain list policy objects |

### Data Sources

| Data Source | Description |
|-------------|-------------|
| `cpwai_workforce_ai_chats_rulebase` | Read the current Workforce AI chats rulebase |
| `cpwai_workforce_ai_access_rulebase` | Read the current Workforce AI access rulebase |
| `cpwai_workforce_ai_agents_rulebase` | Read the current Workforce AI agents rulebase |
| `cpwai_browse_dlp_rulebase` | Read the current Browse DLP rulebase |
| `cpwai_browse_access_rulebase` | Read the current Browse access rulebase |
| `cpwai_browse_secure_browsing_rulebase` | Read the current Browse secure browsing rulebase |

## Documentation

- [Provider Documentation](docs/index.md)
- [Getting Started Guide](docs/guides/getting-started.md)
- [Policy Reference](docs/guides/policy-reference.md)
- [AI Applications Catalog](docs/guides/ai-applications-catalog.md)
- [DLP Data Types](docs/guides/dlp-data-types.md)

## Report Bug

In case of an issue or a bug found in the provider, please open an [issue](https://github.com/CheckPointSW/terraform-provider-checkpoint-workforce-ai/issues).

## Contributors
- Haim Kastner - [haimk@checkpoint.com](mailto:haimk@checkpoint.com)
