# Resource: cpwai_browse_dlp_rule

Manages a DLP rule in the Browse Protect policy. Controls data loss prevention for file uploads, downloads, pastes, and prompts to web-based chat services.

## Example Usage

```hcl
resource "cpwai_browse_dlp_rule" "block_credit_cards" {
  name        = "Block Credit Card Uploads in Browser"
  description = "Prevent uploading files containing credit card numbers via the browser"
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

## Argument Reference

The following arguments are supported:

| Argument | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| `name` | String | Yes | — | Human-readable rule name. Used for display in the Infinity Portal UI and Terraform state. |
| `description` | String | No | `""` | Free-text explanation of the rule's purpose. |
| `order` | Number | No | Computed | Position in the rulebase (0-based). Lower values are evaluated first. |
| `active` | Boolean | No | `true` | Whether the rule is currently enforced. Set to `false` to disable without deleting. |
| `policy` | String (JSON) | Yes | — | The enforcement policy as a JSON string (use `jsonencode({...})`). See the [Policy Reference](../guides/policy-reference.md) for the Chats DLP policy schema. |
| `source` | List | No | — | List of users, groups, or the entire organization this rule applies to (source entries). An empty list means the rule matches all users. See [Source](#source) below. |


### Source

Each element in the `source` list supports:

| Argument | Type | Required | Description |
|----------|------|----------|-------------|
| `assignment_type` | String | Yes | Assignment scope: `ASSIGNMENT_TYPE_USER`, `ASSIGNMENT_TYPE_GROUP`, or `ASSIGNMENT_TYPE_ENTIRE_ORG`. |
| `assignment_id` | String | No | UUID of the user or group. Not required for `ASSIGNMENT_TYPE_ENTIRE_ORG`. |
| `display_name` | String | No | Display name. Computed by the server if not provided. |

### Policy Fields (Chats DLP)

The `policy` JSON object supports the following fields:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `event_type` | String | Yes | Event to apply DLP on: `file_upload`, `file_download`, `paste`, `prompt`, `copy`. |
| `action` | String | Yes | DLP action: `prevent`, `ask`, `redact`, `detect`, `allow`, `block`. **Note:** rules without `data_types` (any-text) can only use `block` or `allow`. |
| `logging` | String | No | `"enabled"` or `"disabled"`. Default: `"enabled"`. |
| `services_and_application` | Object | No | Service scope. See below. |
| `data_types` | Array | No | DLP data types to detect. See below. |

#### services_and_application

| Field | Type | Description |
|-------|------|-------------|
| `mode` | String | `"all"` (apply to all services) or `"selected"` (apply to listed services only). |
| `domain` | Array | List of domains. Each: `{ domain = "example.com", match_mode = "exact_host" \| "any_subdomain" }`. |
| `url` | Array | List of URLs. Each: `{ url = "https://..." }`. |
| `category` | Array | List of URLF categories. Each: `{ category_id = <int> }`. Any category except Generative AI Tools (60531762). See [Supported Categories](#supported-categories). |

~> **Browse constraint:** Browse chats rules cannot use `genai_application` in `services_and_application`. To target specific AI applications, use the Workforce AI chats rule (`cpwai_workforce_ai_chats_rule`) instead.

#### Supported Categories

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

#### data_types

Each element in the `data_types` array:

| Field | Type | Description |
|-------|------|-------------|
| `id` | String (UUID) | Unique identifier of the data type. |
| `name` | String | Display name (e.g. "Credit Card Number"). |
| `type` | String | `"PRE_DEFINED"` or `"CUSTOM"`. |
| `matching_level` | Number | Optional. Matching sensitivity (1-5). |

Data types are managed at the tenant level via the [DLP Datatypes API](https://app.swaggerhub.com/apis/Check-Point/checkpoint-ai-security/1.0.0#/DLP%20Datatypes). See the [DLP Data Types guide](../guides/dlp-data-types.md) for details on listing data types.

## Attribute Reference

In addition to the arguments above, the following attributes are exported:

| Attribute | Type | Description |
|-----------|------|-------------|
| `id` | String | Unique identifier of the rule (UUID). Assigned by the server on creation. |
| `version` | Number | Rule version number, incremented on every modification. |

## Import

An existing Browse DLP rule can be imported using its rule ID:

```shell
terraform import cpwai_browse_dlp_rule.example <rule_id>
```
