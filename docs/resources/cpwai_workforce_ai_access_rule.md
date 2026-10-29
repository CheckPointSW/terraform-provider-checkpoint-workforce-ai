# Resource: cpwai_workforce_ai_access_rule

Manages an Access rule in the Workforce AI policy. Controls which Workforce AI applications users can access.

## Example Usage

```hcl
resource "cpwai_workforce_ai_access_rule" "block_chatgpt" {
  name   = "Block ChatGPT Access"
  order  = 0
  active = true

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

  source = [
    {
      assignment_type = "ASSIGNMENT_TYPE_ENTIRE_ORG"
    }
  ]
}
```

~> **Important:** Workforce AI access rules cannot use `domain` in `services_and_application`. Use `genai_application` instead. The `domain` field is only supported in Browse access rules (`cpwai_browse_access_rule`).

## Argument Reference

The following arguments are supported:

| Argument | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| `name` | String | Yes | — | Human-readable rule name. Used for display in the Infinity Portal UI and Terraform state. |
| `description` | String | No | `""` | Free-text explanation of the rule's purpose. |
| `order` | Number | No | Computed | Position in the rulebase (0-based). Lower values are evaluated first. |
| `active` | Boolean | No | `true` | Whether the rule is currently enforced. Set to `false` to disable without deleting. |
| `policy` | String (JSON) | Yes | — | The enforcement policy as a JSON string (use `jsonencode({...})`). See the [Policy Reference](../guides/policy-reference.md) for the Access policy schema. |
| `source` | List | No | — | List of users, groups, or the entire organization this rule applies to (source entries). An empty list means the rule matches all users. See [Source](#source) below. |


### Source

Each element in the `source` list supports:

| Argument | Type | Required | Description |
|----------|------|----------|-------------|
| `assignment_type` | String | Yes | Assignment scope: `ASSIGNMENT_TYPE_USER`, `ASSIGNMENT_TYPE_GROUP`, or `ASSIGNMENT_TYPE_ENTIRE_ORG`. |
| `assignment_id` | String | No | UUID of the user or group. Not required for `ASSIGNMENT_TYPE_ENTIRE_ORG`. |
| `display_name` | String | No | Display name. Computed by the server if not provided. |

### Policy Fields (Access)

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `action` | String | Yes | Access action: `"block"`, `"allow"`, or `"ask"`. |
| `logging` | String | No | `"enabled"` or `"disabled"`. |
| `services_and_application` | Object | No | Service scope. See below. |
| `download_file_protection` | String | No | Must be `"na"` for Workforce AI rules. |
| `upload_file_protection` | String | No | Must be `"na"` for Workforce AI rules. |

#### services_and_application

| Field | Type | Description |
|-------|------|-------------|
| `mode` | String | Must be `"selected"`. Workforce AI rules reject `"all"`; to match any AI app, set `category = [{ category_id = 60531762 }]`. |
| `genai_application` | Array | List of Workforce AI apps. Each: `{ id = <int>, mode = "all" \| "selected_account", accounts = [{ account_id = "..." }] }`. See the [AI Applications Catalog guide](../guides/ai-applications-catalog.md). |
| `category` | Array | List of URLF categories. Only the **Generative AI Tools** category (`category_id = 60531762`) is allowed — at most one entry. This acts as a wildcard that matches any Workforce AI application. |

~> **Workforce AI constraint:** Workforce AI access rules cannot use `domain` or `url` in `services_and_application`. To target specific domains or URLs, use the Browse access rule (`cpwai_browse_access_rule`) instead.

## Attribute Reference

In addition to the arguments above, the following attributes are exported:

| Attribute | Type | Description |
|-----------|------|-------------|
| `id` | String | Unique identifier of the rule (UUID). Assigned by the server on creation. |
| `version` | Number | Rule version number, incremented on every modification. |

## Import

An existing Workforce AI Access rule can be imported using its rule ID:

```shell
terraform import cpwai_workforce_ai_access_rule.example <rule_id>
```
