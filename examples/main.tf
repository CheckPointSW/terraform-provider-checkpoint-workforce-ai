terraform {
  required_providers {
    cpwai = {
      source = "CheckPointSW/checkpoint-workforce-ai"
      version = "0.1.0"
    }
  }
}

variable "checkpoint_client_id" {
  description = "Client ID from the Infinity Portal API key"
  type        = string
  sensitive   = true
}

variable "checkpoint_access_key" {
  description = "Access key from the Infinity Portal API key"
  type        = string
  sensitive   = true
}

variable "checkpoint_region" {
  description = "Region for the Infinity Portal API (eu or us)"
  type        = string
  default     = "us"
}

provider "cpwai" {
  client_id  = var.checkpoint_client_id
  access_key = var.checkpoint_access_key
  region     = var.checkpoint_region
}

# --- Workforce AI: Block PII uploads ---

resource "cpwai_workforce_ai_chats_rule" "block_pii" {
  name        = "Block PII in Chat Uploads"
  description = "Prevent uploading files containing PII to AI services"
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

# --- Workforce AI: Access control (workforce_ai tag requires genai_application, not domain) ---

resource "cpwai_workforce_ai_access_rule" "restrict_chatgpt" {
  name   = "Restrict ChatGPT Access"
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

# --- Workforce AI: Agents/MCP policy ---

resource "cpwai_workforce_ai_agents_rule" "allow_mcp" {
  name   = "Allow Internal MCP Servers"
  order  = 0
  active = true

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

  source = [
    {
      assignment_type = "ASSIGNMENT_TYPE_ENTIRE_ORG"
    }
  ]
}

# --- Browse Security: Protect internal domains with Secure Browsing ---

resource "cpwai_browse_domains_object" "org_internal_domains" {
  name        = "Organization Internal Domains"
  description = "Internal corporate domains that require browsing protection"

  object_data = jsonencode({
    domains = [
      {
        domain     = "portal.myorg.com"
        match_mode = "exact_host"
      },
      {
        domain     = "apps.myorg.com"
        match_mode = "any_subdomain"
      },
      {
        domain     = "intranet.myorg.com"
        match_mode = "any_subdomain"
      }
    ]
  })
}

resource "cpwai_browse_file_protection_object" "internal_file_policy" {
  name        = "Internal File Protection"
  description = "Emulate downloads and block unsupported uploads on internal domains"

  object_data = jsonencode({
    file_download_action_unsupported = "Block"
    file_upload_action_unsupported   = "Block"
    file_download_size_limit_mb      = 100
    file_upload_size_limit_mb        = 50
    file_download_error_fail_close   = "FailClose"
    file_upload_error_fail_close     = "FailClose"
  })
}

resource "cpwai_browse_secure_browsing_rule" "protect_internal_domains" {
  name        = "Protect Internal Domains"
  description = "Enable secure browsing and file emulation for internal corporate domains"
  order       = 0
  active      = true

  policy = jsonencode({
    password_reuse    = "prevent"
    zero_phishing     = "prevent"
    safe_search       = "on"
    search_reputation = "on"
  })

  source = [
    {
      assignment_type = "ASSIGNMENT_TYPE_ENTIRE_ORG"
    }
  ]

  objects = {
    PROTECTED_DOMAINS = [cpwai_browse_domains_object.org_internal_domains.id]
  }
}

# --- Browse Security: Allow access to internal resources ---

resource "cpwai_browse_access_rule" "allow_internal_access" {
  name        = "Allow Access to Internal Resources"
  description = "Allow employees to access internal corporate domains"
  order       = 1
  active      = true

  policy = jsonencode({
    action  = "allow"
    logging = "enabled"
    services_and_application = { mode = "all" }
  })

  source = [
    {
      assignment_type = "ASSIGNMENT_TYPE_ENTIRE_ORG"
    }
  ]

  objects = {
    DOMAIN_SERVICES = [cpwai_browse_domains_object.org_internal_domains.id]
    FILE_EMULATION  = [cpwai_browse_file_protection_object.internal_file_policy.id]
  }
}

# --- Data Sources ---

data "cpwai_workforce_ai_chats_rulebase" "current" {}

output "chats_rulebase_version" {
  value = data.cpwai_workforce_ai_chats_rulebase.current.rulebase_version
}
