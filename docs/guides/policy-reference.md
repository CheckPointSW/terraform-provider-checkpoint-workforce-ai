# Policy Reference

Each rule resource has a `policy` attribute that accepts a JSON string (use `jsonencode({...})`). The structure varies by rule type. This document describes the complete schema for each.

---

## Chats Rules (DLP Policy)

Used by: `cpwai_workforce_ai_chats_rule`, `cpwai_browse_dlp_rule`

Chats rules apply Data Loss Prevention (DLP) policies to chat/messaging events.

### Workforce AI Chats Example

```hcl
policy = jsonencode({
  event_type = "file_upload"
  action     = "prevent"
  logging    = "enabled"
  services_and_application = {
    mode = "selected"
    genai_application = [
      { id = 1, mode = "all" }
    ]
  }
  data_types = [
    {
      id   = "cf0523c1-537e-4a4b-8bb8-084b7b9e0b45"
      name = "Credit Card Number"
      type = "PRE_DEFINED"
    }
  ]
})
```

### Browse Chats Example

```hcl
policy = jsonencode({
  event_type = "file_upload"
  action     = "prevent"
  logging    = "enabled"
  services_and_application = {
    mode = "selected"
    domain = [
      { domain = "chat.example.com", match_mode = "any_subdomain" }
    ]
  }
  data_types = [
    {
      id   = "cf0523c1-537e-4a4b-8bb8-084b7b9e0b45"
      name = "Credit Card Number"
      type = "PRE_DEFINED"
    }
  ]
})
```

### Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `event_type` | string | Yes | Event to apply DLP on. See [Event Types](#event-types). |
| `action` | string | Yes | Action when DLP detects a match. See [DLP Actions](#dlp-actions). |
| `logging` | string | No | `"enabled"` or `"disabled"`. Default: `"enabled"`. |
| `services_and_application` | object | No | Scope of services. See [Services and Application](#services-and-application). |
| `data_types` | array | No | DLP data types to detect. See [Data Types](#data-types-dlp) and the [DLP Data Types guide](dlp-data-types.md). |

### Event Types

| Value | Description |
|-------|-------------|
| `file_upload` | File upload to a service |
| `file_download` | File download from a service |
| `paste` | Text pasted into a service |
| `prompt` | Text prompt sent to an AI service |
| `copy` | Text copied from a service |

### DLP Actions

| Value | Description |
|-------|-------------|
| `prevent` | Block the action entirely |
| `ask` | Prompt the user for confirmation |
| `redact` | Redact the sensitive data and allow |
| `detect` | Allow but log the detection |
| `allow` | Allow without restriction |
| `block` | Block the action entirely |

~> **Note:** Rules without `data_types` (any-text matching) can only use `block` or `allow` actions. The `detect`, `prevent`, `ask`, and `redact` actions require at least one data type to be specified.

---

## Access Rules

Used by: `cpwai_workforce_ai_access_rule`, `cpwai_browse_access_rule`

Access rules control which services and applications users can access.

### Workforce AI Access Example

```hcl
policy = jsonencode({
  action  = "block"
  logging = "enabled"
  services_and_application = {
    mode = "selected"
    genai_application = [
      {
        id   = 1
        mode = "all"
      }
    ]
  }
})
```

### Browse Access Example

```hcl
policy = jsonencode({
  action                   = "allow"
  logging                  = "enabled"
  download_file_protection = "WaitForEmulation"
  upload_file_protection   = "Detect"
  services_and_application = {
    mode = "selected"
    domain = [
      { domain = "portal.myorg.com", match_mode = "exact_host" }
    ]
  }
})
```

### Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `action` | string | Yes | Action for access control: `"block"`, `"allow"`, or `"ask"`. |
| `logging` | string | No | `"enabled"` or `"disabled"`. |
| `services_and_application` | object | No | Scope of services. See [Services and Application](#services-and-application). |
| `download_file_protection` | string | No | File download protection action. Workforce AI rules: must be `"na"`. Browse rules: must not be `"na"` — valid values: `"Block"`, `"Allow"`, `"EmulateInBackground"`, `"WaitForEmulation"`, `"Detect"`, `"Extract"`. |
| `upload_file_protection` | string | No | File upload protection action. Workforce AI rules: must be `"na"`. Browse rules: must not be `"na"` — valid values: `"Block"`, `"Allow"`, `"EmulateInBackground"`, `"WaitForEmulation"`, `"Detect"`. |

---

## Agents Rules (MCP Server Policy)

Used by: `cpwai_workforce_ai_agents_rule`

Agents rules control AI agents and MCP (Model Context Protocol) server interactions.

```hcl
policy = jsonencode({
  action  = "allow"
  logging = "enabled"
  clients = { mode = "all" }
  servers = {
    mcp_servers_mode = "manual"
    manual = [
      {
        server_type = "remote_server"
        remote_server = {
          match_mode = "domain"
          domain = {
            domain_name = "mcp.internal.com"
            match_mode  = "exact_host"
          }
        }
      }
    ]
  }
  tooling = { match_mode = "all" }
})
```

### Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `action` | string | Yes | Action: `"block"`, `"allow"`, or `"ask"`. |
| `logging` | string | No | `"enabled"` or `"disabled"`. |
| `clients` | object | No | MCP client scope. |
| `servers` | object | No | MCP server scope. |
| `tooling` | object | No | Tool matching configuration. |

### Clients Object

| Field | Type | Description |
|-------|------|-------------|
| `mode` | string | `"all"` or `"selected"` |
| `selected` | array | List of MCP clients when `mode = "selected"`. Each: `{ os, name_id }` |

### Servers Object

| Field | Type | Description |
|-------|------|-------------|
| `mcp_servers_mode` | string | `"all"`, `"risk_level"`, `"manual"`, or `"managed"` |
| `risk_level` | object | Risk level configuration (when `mcp_servers_mode = "risk_level"`) |
| `managed` | object | Managed MCP configuration (when `mcp_servers_mode = "managed"`). Contains `name_ids` array |
| `manual` | array | List of MCP server definitions (when `mcp_servers_mode = "manual"`). Each has `server_type` and a type-specific config |

Server types: `"all"`, `"node_server"`, `"python_server"`, `"claude_extensions"`, `"executable_server"`, `"unknown_server"`, `"docker_gateway"`, `"docker_runner"`, `"remote_server"`.

For `remote_server`, the config has `match_mode` (`"url"` or `"domain"`), and either a `domain` object (`{ domain_name, match_mode }`) or a `url` object (`{ url, match_mode }`).

### Tooling Object

| Field | Type | Description |
|-------|------|-------------|
| `match_mode` | string | `"tools_include"`, `"tools_exclude"`, `"operations"`, or `"all"` |
| `tools` | array | List of tool items. Each: `{ tool_name, match_mode = "exact" \| "contains" }` |
| `operations` | object | Server operations to match. Keys: `CREATE`, `READ`, `UPDATE`, `DELETE`, `OTHER`. Values: `"match"` or `"unmatch"` |

---

## Secure Browsing Rules

Used by: `cpwai_browse_secure_browsing_rule`

Secure browsing rules define threat-prevention policies for web browsing: phishing protection, password-reuse detection, safe search, and search reputation.

```hcl
policy = jsonencode({
  password_reuse    = "prevent"
  zero_phishing     = "detect"
  safe_search       = "on"
  search_reputation = "on"
  protected_domains = [
    {
      domain     = "portal.myorg.com"
      match_mode = "exact_host"
    }
  ]
})
```

### Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `password_reuse` | string | No | Password reuse detection: `"detect"`, `"prevent"`, `"ask"`, or `"off"`. |
| `zero_phishing` | string | No | Phishing protection: `"detect"`, `"prevent"`, `"ask"`, or `"off"`. |
| `safe_search` | string | No | Safe search enforcement: `"on"` or `"off"`. |
| `search_reputation` | string | No | Search reputation filtering: `"on"` or `"off"`. |
| `domain_exclusions` | array | No | Domains to exclude. Each: `{ domain, match_mode }`. |
| `protected_domains` | array | No | Domains to protect. Each: `{ domain, match_mode }`. |

---

## Services and Application

The `services_and_application` object defines which services/applications a rule applies to. Used in Chats and Access rules.

The available fields depend on whether the rule is a **Workforce AI** rule or a **Browse** rule:

| Field | Workforce AI rules                                        | Browse rules                |
|-------|----------------------------------------------------|-----------------------------|
| `domain` | Not allowed                                     | Allowed                     |
| `url` | Not allowed                                        | Allowed                     |
| `genai_application` | Allowed                              | Not allowed                 |
| `category` | Only `60531762` act as any generative AI tools| Any non `60531762` category |

### Workforce AI Example

```hcl
services_and_application = {
  mode = "selected"

  # Workforce AI Applications (by ID)
  genai_application = [
    {
      id   = 1
      mode = "all"  # or "selected_account"
      accounts = [  # only when mode = "selected_account"
        { account_id = "org-123" }
      ]
    }
  ]
}
```

### Browse Example

```hcl
services_and_application = {
  mode = "selected"

  # Domains
  domain = [
    {
      domain     = "chat.openai.com"
      match_mode = "any_subdomain"  # or "exact_host" (default)
    }
  ]

  # URLs
  url = [
    { url = "https://example.com/api" }
  ]

  # URLF Categories (by predefined ID)
  category = [
    { category_id = 1 }
  ]
}
```

### Fields

| Field | Type | Description |
|-------|------|-------------|
| `mode` | string | `"all"` — apply to all services; `"selected"` — apply to listed services only |
| `domain` | array | **Browse only.** List of domains with `domain` (string) and optional `match_mode` |
| `url` | array | **Browse only.** List of URLs with `url` (string) |
| `genai_application` | array | **GenAI only.** List of Workforce AI app IDs with optional account selection. See the [AI Applications Catalog guide](ai-applications-catalog.md). |
| `category` | array | List of URLF category IDs. Workforce AI rules: only `60531762` (Generative AI Tools) allowed. Browse rules: any category (see [Supported Categories](#supported-categories)). |

### Domain Match Modes

| Value | Description |
|-------|-------------|
| `exact_host` | Matches the exact hostname (default) |
| `any_subdomain` | Matches any subdomain (e.g. `example.com` matches `sub.example.com`) |

### Account Selection Modes

| Value | Description |
|-------|-------------|
| `all` | Policy applies to all tool instances regardless of account (default) |
| `selected_account` | Policy applies only to instances connected to specified account IDs |

### Supported Categories

Common URLF categories that can be used in Browse rules:

| Category | `category_id` |
|----------|---------------|
| Uncategorized | `0` |
| Computers & Internet | `1` |
| Education | `3` |
| Entertainment | `5` |
| Financial Services | `7` |
| Gambling | `9` |
| Government / Military | `11` |
| Hacking | `13` |
| Health | `15` |
| Illegal / Questionable | `17` |
| Job Search / Careers | `21` |
| News / Media | `24` |
| Phishing | `31` |
| Pornography | `33` |
| Real Estate | `34` |
| Shopping | `42` |
| Business / Economy | `51` |
| Sports | `54` |
| Social Networking | `52000069` |
| File Storage / Sharing | `52000136` |
| Search Engines / Portals | `52000132` |
| Artificial Intelligence (AI) | `60517663` |

---

## Data Types (DLP)

Data types define the sensitive information patterns that DLP (Data Loss Prevention) rules detect in Chats rules. Each data type represents a category of sensitive data — such as credit card numbers, social security numbers, or custom patterns defined by your organization.

### Using Data Types in Rules

In Chats DLP rules, reference data types in the `data_types` array of the policy:

```hcl
data_types = [
  {
    id             = "cf0523c1-537e-4a4b-8bb8-084b7b9e0b45"
    name           = "Credit Card Number"
    type           = "PRE_DEFINED"
    matching_level = 3  # optional: sensitivity 1-5
  }
]
```

| Field | Type | Description |
|-------|------|-------------|
| `id` | string (UUID) | Unique identifier of the data type |
| `name` | string | Display name |
| `type` | string | `"PRE_DEFINED"` or `"CUSTOM"` |
| `matching_level` | integer | Optional. Matching sensitivity level |

Use the `cpwai_workforce_ai_chats_rulebase` or `cpwai_browse_dlp_rulebase` data source to discover available predefined data types. 

For the full API reference, see the [DLP Datatypes API documentation](https://app.swaggerhub.com/apis/Check-Point/checkpoint-ai-security/1.0.0#/DLP%20Datatypes).

---

## Common Enumerations

### Logging Status

| Value | Description |
|-------|-------------|
| `enabled` | Log events matching this rule |
| `disabled` | Do not log |

### Assignment Types

| Value | Description |
|-------|-------------|
| `ASSIGNMENT_TYPE_USER` | Individual user |
| `ASSIGNMENT_TYPE_GROUP` | User group |
| `ASSIGNMENT_TYPE_ENTIRE_ORG` | Entire organization |
