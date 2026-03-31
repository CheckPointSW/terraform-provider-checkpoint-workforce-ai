# Getting Started with Check Point Workforce AI Provider

This guide walks you through setting up the provider and creating your first security policies.

## Prerequisites

- [Terraform](https://developer.hashicorp.com/terraform/install) >= 1.0
- A Check Point Infinity Portal account with **Workforce AI** or **Browse Security** license
- An API key (Client ID + Access Key)

## Step 1: Create an API Key

1. Log in to the [Infinity Portal](https://portal.checkpoint.com)
2. Navigate to **Settings → API Keys**
3. Click **New → New Account API Key**
4. In the **Service** dropdown, select:
   - **Workforce AI Security** — for AI Security resources (`cpwai_workforce_ai_*`)
   - **Browser Security** — for Browse Security resources (`cpwai_browse_*`)
5. Copy the **Client ID** and **Access Key**

## Step 2: Configure Environment Variables

```bash
export TF_VAR_checkpoint_client_id="your-client-id"
export TF_VAR_checkpoint_access_key="your-access-key"
export TF_VAR_checkpoint_region="us"  # or "eu"
```

## Step 3: Write Your Configuration

Create a `main.tf` file:

```hcl
terraform {
  required_providers {
    cpwai = {
      source  = "CheckPointSW/checkpoint-workforce-ai"
      version = "~> 1.0"
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

variable "checkpoint_region" {
  type    = string
  default = "us"
}

provider "cpwai" {
  client_id  = var.checkpoint_client_id
  access_key = var.checkpoint_access_key
  region     = var.checkpoint_region
}

# Block uploading credit card numbers to AI chat services
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

## Step 4: Apply

```bash
terraform init
terraform plan
terraform apply
```

## Step 5: Verify

Check the rulebase version using a data source:

```hcl
data "cpwai_workforce_ai_chats_rulebase" "current" {}

output "rulebase_version" {
  value = data.cpwai_workforce_ai_chats_rulebase.current.rulebase_version
}
```

## Next Steps

- See the [Policy Reference](policy-reference.md) for the full JSON structure of each rule type
- See [AI Applications Catalog](ai-applications-catalog.md) for looking up application IDs
- See [DLP Data Types](dlp-data-types.md) for looking and managing data loss prevention patterns
- Browse the [resource documentation](../resources/cpwai_workforce_ai_chats_rule) for all available resources
- Use `terraform import cpwai_workforce_ai_chats_rule.name <rule_id>` to import existing rules
