package sdk

import (
	"encoding/json"
	"time"
)

// --- Common Models ---

// Rule represents a policy rule in any rulebase.
type Rule struct {
	RuleID      string                 `json:"rule_id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Version     int                    `json:"version"`
	Order       int                    `json:"order"`
	Active      bool                   `json:"active"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	Policy      map[string]interface{} `json:"policy"`
	Source      []Assignment           `json:"source"`
	Tags        []string               `json:"tags,omitempty"`
	Objects     map[string][]string    `json:"objects,omitempty"`
}

// Rulebase represents a collection of rules with a version.
type Rulebase struct {
	RulebaseVersion int    `json:"rulebase_version"`
	Rules           []Rule `json:"rules"`
}

// Assignment represents a user/group source entry on a rule.
type Assignment struct {
	AssignmentID   string `json:"assignment_id,omitempty"`
	DisplayName    string `json:"display_name,omitempty"`
	AssignmentType string `json:"assignment_type"`
}

// --- Services and Application ---

// ServicesAndApplication configures which services/apps a rule applies to.
type ServicesAndApplication struct {
	Mode            *string           `json:"mode,omitempty"`
	Domain          []DomainItem      `json:"domain,omitempty"`
	URL             []URLItem         `json:"url,omitempty"`
	GenAIApp        []GenAIAppItem    `json:"genai_application,omitempty"`
	Category        []CategoryItem    `json:"category,omitempty"`
}

// DomainItem represents a domain in a service/application config.
type DomainItem struct {
	Domain    string  `json:"domain"`
	MatchMode *string `json:"match_mode,omitempty"` // "exact_host" or "any_subdomain"
}

// URLItem represents a URL entry.
type URLItem struct {
	URL string `json:"url"`
}

// GenAIAppItem represents a GenAI application entry.
type GenAIAppItem struct {
	ID       int            `json:"id"`
	Mode     *string        `json:"mode,omitempty"` // "all" or "selected_account"
	Accounts []AccountItem  `json:"accounts,omitempty"`
}

// AccountItem represents a selected account.
type AccountItem struct {
	AccountID string `json:"account_id"`
}

// CategoryItem represents a URLF category.
type CategoryItem struct {
	CategoryID int `json:"category_id"`
}

// --- DLP Policy (Chats rules) ---

// DLPPolicy represents the policy for Chats/DLP rules.
type DLPPolicy struct {
	EventType              *string                 `json:"event_type,omitempty"`
	ServicesAndApplication *ServicesAndApplication  `json:"services_and_application,omitempty"`
	DataTypes              []DLPDataTypeRef        `json:"data_types,omitempty"`
	Action                 *string                 `json:"action,omitempty"`
	Logging                *string                 `json:"logging,omitempty"`
}

// DLPDataTypeRef references a DLP data type in a policy.
type DLPDataTypeRef struct {
	ID            string `json:"id"`
	Name          string `json:"name,omitempty"`
	Type          string `json:"type,omitempty"`          // "PRE_DEFINED" or "CUSTOM"
	MatchingLevel *int   `json:"matchingLevel,omitempty"`
}

// --- Access Policy ---

// AccessPolicy represents the policy for Access rules.
type AccessPolicy struct {
	ServicesAndApplication *ServicesAndApplication `json:"services_and_application,omitempty"`
	Action                 *string                `json:"action,omitempty"` // "block", "allow", "ask"
	Logging                *string                `json:"logging,omitempty"`
}

// --- MCP/Agents Policy ---

// AgentsPolicy represents the policy for Agents/MCP rules.
type AgentsPolicy struct {
	Action  *string        `json:"action,omitempty"` // "block", "allow"
	Logging *string        `json:"logging,omitempty"`
	Clients *ClientsConfig `json:"clients,omitempty"`
	Servers *ServersConfig `json:"servers,omitempty"`
	Tooling *ToolingConfig `json:"tooling,omitempty"`
}

// ClientsConfig configures MCP client selection.
type ClientsConfig struct {
	Mode     *string     `json:"mode,omitempty"` // "all" or "selected"
	Selected []MCPClient `json:"selected,omitempty"`
}

// MCPClient represents a selected MCP client.
type MCPClient struct {
	OS     *string `json:"os,omitempty"`      // "windows", "macos", "linux", "all"
	NameID *string `json:"name_id,omitempty"` // e.g. "claude_desktop", "docker_desktop"
}

// ServersConfig configures MCP server selection.
type ServersConfig struct {
	MCPServersMode *string            `json:"mcp_servers_mode,omitempty"` // "all", "risk_level", "manual", "managed"
	RiskLevel      map[string]interface{} `json:"risk_level,omitempty"`
	Managed        map[string]interface{} `json:"managed,omitempty"`
	Manual         []map[string]interface{} `json:"manual,omitempty"`
}

// ToolingConfig configures tool matching.
type ToolingConfig struct {
	MatchMode  *string                `json:"match_mode,omitempty"` // "tools_include", "tools_exclude", "operations", "all"
	Tools      []MCPToolItem          `json:"tools,omitempty"`
	Operations map[string]string      `json:"operations,omitempty"` // CREATE/READ/UPDATE/DELETE/OTHER -> "match"/"unmatch"
}

// MCPToolItem represents a tool in the tooling config.
type MCPToolItem struct {
	ToolName  *string `json:"tool_name,omitempty"`
	MatchMode *string `json:"match_mode,omitempty"` // "exact" or "contains"
}

// --- Secure Browsing Policy (Browse only) ---

// SecureBrowsingPolicy represents the policy for Secure Browsing rules.
type SecureBrowsingPolicy struct {
	PasswordReuse    *string      `json:"password_reuse,omitempty"`    // "detect", "prevent", "ask", "off"
	ZeroPhishing     *string      `json:"zero_phishing,omitempty"`    // "detect", "prevent", "ask", "off"
	SafeSearch       *string      `json:"safe_search,omitempty"`      // "on", "off"
	SearchReputation *string      `json:"search_reputation,omitempty"` // "on", "off"
	DomainExclusions []DomainItem `json:"domain_exclusions,omitempty"`
	ProtectedDomains []DomainItem `json:"protected_domains,omitempty"`
}

// --- Policy Objects (Browse only) ---

// FileProtectionObject represents a file protection policy object.
type FileProtectionObject struct {
	ObjectID    string           `json:"object_id,omitempty"`
	Name        string           `json:"name"`
	Description string           `json:"description,omitempty"`
	ObjectData  json.RawMessage  `json:"object_data"`
	CreatedAt   *time.Time       `json:"created_at,omitempty"`
	UpdatedAt   *time.Time       `json:"updated_at,omitempty"`
}

// FileProtectionObjectData holds the file protection configuration.
type FileProtectionObjectData struct {
	FileDownloadActionUnsupported      *string           `json:"file_download_action_unsupported,omitempty"`
	FileUploadActionUnsupported        *string           `json:"file_upload_action_unsupported,omitempty"`
	FileDownloadSizeLimitMB            *int              `json:"file_download_size_limit_mb,omitempty"`
	FileUploadSizeLimitMB              *int              `json:"file_upload_size_limit_mb,omitempty"`
	FileDownloadErrorFailClose         *string           `json:"file_download_error_fail_close,omitempty"`
	FileDownloadSizeLimitFailClose     *string           `json:"file_download_size_limit_fail_close,omitempty"`
	FileUploadErrorFailClose           *string           `json:"file_upload_error_fail_close,omitempty"`
	FileUploadSizeLimitFailClose       *string           `json:"file_upload_size_limit_fail_close,omitempty"`
	FileDownloadPerSupportedFile       map[string]string `json:"file_download_action_per_supported_file,omitempty"`
	FileDownloadPerUnsupportedFile     map[string]string `json:"file_download_action_per_unsupported_file,omitempty"`
	FileUploadPerSupportedFile         map[string]string `json:"file_upload_action_per_supported_file,omitempty"`
	FileUploadPerUnsupportedFile       map[string]string `json:"file_upload_action_per_unsupported_file,omitempty"`
}

// DomainsObject represents a domains policy object.
type DomainsObject struct {
	ObjectID    string          `json:"object_id,omitempty"`
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	ObjectData  json.RawMessage `json:"object_data"`
	CreatedAt   *time.Time      `json:"created_at,omitempty"`
	UpdatedAt   *time.Time        `json:"updated_at,omitempty"`
}

// DomainsObjectData holds domain list configuration.
type DomainsObjectData struct {
	Domains []DomainItem `json:"domains"`
}

// --- DLP Data Types ---

// DLPDataType represents a DLP data type definition.
type DLPDataType struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Type          string `json:"type"` // "PRE_DEFINED" or "CUSTOM"
	Description   string `json:"description,omitempty"`
	MatchingLevel *int   `json:"matchingLevel,omitempty"`
}

// TenantDLPConfig represents tenant-level DLP data type configuration.
type TenantDLPConfig struct {
	DataTypes []DLPDataType `json:"data_types"`
}

// --- Request Bodies ---

// AddRuleRequest is the request body for creating any rule type.
type AddRuleRequest struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Order       int                    `json:"order"`
	Policy      map[string]interface{} `json:"policy,omitempty"`
	Source      []Assignment           `json:"source,omitempty"`
}

// SetPolicyRequest replaces the entire policy of a rule.
type SetPolicyRequest struct {
	RuleID  string                 `json:"rule_id"`
	Version int                    `json:"version"`
	Policy  map[string]interface{} `json:"policy"`
}

// PatchPolicyRequest partially updates a rule's policy (deep merge).
type PatchPolicyRequest struct {
	RuleID  string                 `json:"rule_id"`
	Version int                    `json:"version"`
	Policy  map[string]interface{} `json:"policy"`
}

// SetInfoRequest updates rule display info.
type SetInfoRequest struct {
	RuleID      string `json:"rule_id"`
	Version     int    `json:"version"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// ReorderRequest changes a rule's position in the rulebase.
type ReorderRequest struct {
	RuleID   string `json:"rule_id"`
	Version  int    `json:"version"`
	NewOrder int    `json:"new_order"`
}

// SetSourceRequest replaces the source (user/group assignments) of a rule.
type SetSourceRequest struct {
	RuleID  string       `json:"rule_id"`
	Source  []Assignment `json:"source"`
}

// SetActiveRequest enables or disables a rule.
type SetActiveRequest struct {
	RuleID  string `json:"rule_id"`
	Version int    `json:"version"`
	Active  bool   `json:"active"`
}

// SetObjectsRequest sets (replaces) all objects for a rule for a given feature.
type SetObjectsRequest struct {
	RuleID    string   `json:"rule_id"`
	Feature   string   `json:"feature"`
	ObjectIDs []string `json:"object_ids"`
}

// DeleteObjectRequest deletes a policy object.
type DeleteObjectRequest struct {
	ObjectID string `json:"object_id"`
}

// CreateObjectResponse is the API response after creating an object.
type CreateObjectResponse struct {
	ObjectID string `json:"object_id"`
}

// ObjectResponseItem is metadata for a single object in an update response.
type ObjectResponseItem struct {
	ObjectID string `json:"object_id"`
	Name     string `json:"name"`
	Version  int    `json:"version"`
}

// UpdateObjectRequest is the API request for updating an object.
// The object_id identifies which object to update; other fields are optional.
type UpdateObjectRequest struct {
	ObjectID    string      `json:"object_id"`
	Name        string      `json:"name,omitempty"`
	Description string      `json:"description,omitempty"`
	ObjectData  interface{} `json:"object_data,omitempty"`
}

// UpdateObjectResponse is the API response after updating an object.
type UpdateObjectResponse struct {
	Objects []ObjectResponseItem `json:"objects"`
}
