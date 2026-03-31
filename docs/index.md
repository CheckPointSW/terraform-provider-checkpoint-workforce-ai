# Check Point Workforce AI Provider
The `checkpoint-workforce-ai` provider manages security policies for **Workforce AI** and **Browse Security** products via the Check Point Infinity Portal API.

## Example Usage

```hcl
terraform {
  required_providers {
    cpwai = {
      source  = "CheckPointSW/checkpoint-workforce-ai"
      version = "~> 1.0"
    }
  }
}

provider "cpwai" {
  client_id  = var.checkpoint_client_id
  access_key = var.checkpoint_access_key
  region     = "us"
}
```

## Configuration

| Attribute | Env Var | Required | Default | Description |
|-----------|---------|----------|---------|-------------|
| `client_id` | `CHECKPOINT_INFINITY_CLIENT_ID` | Yes | — | Client ID from the Infinity Portal API key |
| `access_key` | `CHECKPOINT_INFINITY_ACCESS_KEY` | Yes | — | Access key (secret) from the Infinity Portal API key |
| `region` | `CHECKPOINT_INFINITY_REGION` | No | `eu` | API region: `eu` or `us` |

## Resources

### Workforce AI

| Resource | Description |
|----------|-------------|
| [cpwai_workforce_ai_chats_rule](resources/cpwai_workforce_ai_chats_rule.md) | DLP rules for AI chat services |
| [cpwai_workforce_ai_access_rule](resources/cpwai_workforce_ai_access_rule.md) | Access control rules for AI services |
| [cpwai_workforce_ai_agents_rule](resources/cpwai_workforce_ai_agents_rule.md) | Policy rules for AI agents and MCP servers |

### Browse Protect

| Resource | Description |
|----------|-------------|
| [cpwai_browse_dlp_rule](resources/cpwai_browse_dlp_rule.md) | DLP rules for web browsing services |
| [cpwai_browse_access_rule](resources/cpwai_browse_access_rule.md) | Access control rules for web browsing |
| [cpwai_browse_secure_browsing_rule](resources/cpwai_browse_secure_browsing_rule.md) | Secure browsing rules (phishing, password reuse) |
| [cpwai_browse_file_protection_object](resources/cpwai_browse_file_protection_object.md) | Reusable file protection policy objects |
| [cpwai_browse_domains_object](resources/cpwai_browse_domains_object.md) | Reusable domain list policy objects |

## Data Sources

| Data Source | Description |
|-------------|-------------|
| `cpwai_workforce_ai_chats_rulebase` | Read the full Workforce AI Chats rulebase |
| `cpwai_workforce_ai_access_rulebase` | Read the full Workforce AI Access rulebase |
| `cpwai_workforce_ai_agents_rulebase` | Read the full Workforce AI Agents rulebase |
| `cpwai_browse_dlp_rulebase` | Read the full Browse DLP rulebase |
| `cpwai_browse_access_rulebase` | Read the full Browse Access rulebase |
| `cpwai_browse_secure_browsing_rulebase` | Read the full Browse Secure Browsing rulebase |

## Guides

- [Getting Started](guides/getting-started.md)
- [Policy Reference](guides/policy-reference.md)
