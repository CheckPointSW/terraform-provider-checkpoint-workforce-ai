# Resource: cpwai_workforce_ai_agents_rule

Manages an Agents/MCP rule in the Workforce AI policy. Controls AI agent and MCP server interactions.

## Example Usage

```hcl
resource "cpwai_workforce_ai_agents_rule" "allow_internal_mcp" {
  name   = "Allow Internal MCP Servers"
  order  = 0
  active = true

  policy = jsonencode({
    action  = "allow"
    logging = "enabled"
    clients = {
      mode = "all"
    }
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
    tooling = {
      match_mode = "all"
    }
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
| `policy` | String (JSON) | Yes | — | The enforcement policy as a JSON string (use `jsonencode({...})`). See the [Policy Reference](../guides/policy-reference.md) for the Agents/MCP policy schema. |
| `source` | List | No | — | List of users, groups, or the entire organization this rule applies to (source entries). An empty list means the rule matches all users. See [Source](#source) below. |


### Source

Each element in the `source` list supports:

| Argument | Type | Required | Description |
|----------|------|----------|-------------|
| `assignment_type` | String | Yes | Assignment scope: `ASSIGNMENT_TYPE_USER`, `ASSIGNMENT_TYPE_GROUP`, or `ASSIGNMENT_TYPE_ENTIRE_ORG`. |
| `assignment_id` | String | No | UUID of the user or group. Not required for `ASSIGNMENT_TYPE_ENTIRE_ORG`. |
| `display_name` | String | No | Display name. Computed by the server if not provided. |

### Policy Fields (Agents/MCP)

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `action` | String | Yes | Action: `"block"` or `"allow"`. |
| `logging` | String | No | `"enabled"` or `"disabled"`. |
| `clients` | Object | No | MCP client scope. See [clients](#clients) below. |
| `servers` | Object | No | MCP server scope. See [servers](#servers) below. |
| `tooling` | Object | No | Tool matching configuration. See [tooling](#tooling) below. |

#### clients

| Field | Type | Description |
|-------|------|-------------|
| `mode` | String | `"all"` or `"selected"`. |
| `selected` | Array | List of MCP clients when `mode = "selected"`. Each: `{ os = "windows" \| "macos" \| "linux", name_id = "claude_desktop" \| "docker_desktop" \| ... }`. |

#### servers

| Field | Type | Description |
|-------|------|-------------|
| `mcp_servers_mode` | String | Server selection mode: `"all"`, `"risk_level"`, `"manual"`, or `"managed"`. |
| `risk_level` | Object | Risk level configuration (when `mcp_servers_mode = "risk_level"`). |
| `managed` | Object | Managed MCP configuration (when `mcp_servers_mode = "managed"`). Contains `name_ids` array. |
| `manual` | Array | List of MCP server definitions (when `mcp_servers_mode = "manual"`). Each has `server_type` and a type-specific config object. |

Server types: `"all"`, `"node_server"`, `"python_server"`, `"claude_extensions"`, `"executable_server"`, `"unknown_server"`, `"docker_gateway"`, `"docker_runner"`, `"remote_server"`.

For `remote_server`, the config has `match_mode` (`"url"` or `"domain"`), and either a `domain` object (`{ domain_name = "...", match_mode = "exact_host" \| "any_subdomain" }`) or a `url` object (`{ url = "https://...", match_mode = "exact_url" \| "url_contains" }`).

#### tooling

| Field | Type | Description |
|-------|------|-------------|
| `match_mode` | String | `"tools_include"`, `"tools_exclude"`, `"operations"`, or `"all"`. |
| `tools` | Array | List of tool items (when `match_mode` is `"tools_include"` or `"tools_exclude"`). Each: `{ tool_name = "...", match_mode = "exact" \| "contains" }`. |
| `operations` | Object | Server operations to match (when `match_mode = "operations"`). Keys: `CREATE`, `READ`, `UPDATE`, `DELETE`, `OTHER`. Values: `"match"` or `"unmatch"`. |

## Attribute Reference

In addition to the arguments above, the following attributes are exported:

| Attribute | Type | Description |
|-----------|------|-------------|
| `id` | String | Unique identifier of the rule (UUID). Assigned by the server on creation. |
| `version` | Number | Rule version number, incremented on every modification. |

## Import

An existing Workforce AI Agents/MCP rule can be imported using its rule ID:

```shell
terraform import cpwai_workforce_ai_agents_rule.example <rule_id>
```
