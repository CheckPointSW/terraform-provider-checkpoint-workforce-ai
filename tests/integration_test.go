package tests

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/CheckPointSW/terraform-provider-checkpoint-workforce-ai/internal/sdk"
	"github.com/CheckPointSW/terraform-provider-checkpoint-workforce-ai/internal/sdk/auth"
	"github.com/CheckPointSW/terraform-provider-checkpoint-workforce-ai/internal/sdk/browse"
	"github.com/CheckPointSW/terraform-provider-checkpoint-workforce-ai/internal/sdk/workforceai"
)

// ---------- helpers ----------

func loadEnv(t *testing.T) {
	t.Helper()
	// If env vars are already set (e.g. in CI), skip loading .env file
	if os.Getenv("CHECKPOINT_INFINITY_CLIENT_ID") != "" {
		return
	}

	// Walk up from the test file to find the project root .env
	_, filename, _, _ := runtime.Caller(0)
	root := filepath.Dir(filepath.Dir(filename))
	envPath := filepath.Join(root, ".env")

	f, err := os.Open(envPath)
	if err != nil {
		t.Skipf("skipping: no .env file and no env vars set (%s)", envPath)
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		os.Setenv(strings.TrimSpace(k), strings.TrimSpace(v))
	}
}

func envOrFatal(t *testing.T, key string) string {
	t.Helper()
	v := os.Getenv(key)
	if v == "" {
		t.Fatalf("required env var %s is not set", key)
	}
	return v
}

type testClients struct {
	browse *browse.Client
	workforceai *workforceai.Client
}

func newClients(t *testing.T) testClients {
	t.Helper()
	loadEnv(t)

	clientID := envOrFatal(t, "CHECKPOINT_INFINITY_CLIENT_ID")
	accessKey := envOrFatal(t, "CHECKPOINT_INFINITY_ACCESS_KEY")
	region := envOrFatal(t, "CHECKPOINT_INFINITY_REGION")

	var baseURL string
	switch region {
	case "us":
		baseURL = "https://cloudinfra-gw-us.portal.checkpoint.com"
	default:
		baseURL = "https://cloudinfra-gw.portal.checkpoint.com"
	}

	creds := auth.Credentials{ClientID: clientID, AccessKey: accessKey}
	authClient := auth.NewClient(baseURL)
	ctx := context.Background()

	token, err := authClient.GetToken(ctx, creds)
	if err != nil {
		t.Fatalf("auth failed: %v", err)
	}

	return testClients{
		browse: browse.NewClient(baseURL, token, authClient, creds),
		workforceai: workforceai.NewClient(baseURL, token, authClient, creds),
	}
}

// ---------- cleanup ----------

func cleanupAll(t *testing.T, c testClients) {
	t.Helper()
	ctx := context.Background()

	// WorkforceAI rulebases
	for _, getRB := range []func(context.Context) (*sdk.Rulebase, error){
		c.workforceai.GetChatsRulebase,
		c.workforceai.GetAccessRulebase,
		c.workforceai.GetAgentsRulebase,
	} {
		rb, err := getRB(ctx)
		if err != nil {
			t.Logf("  skip rulebase (get failed): %v", err)
			continue
		}
		for _, r := range rb.Rules {
			t.Logf("  deleting workforce_ai rule %s (%s)", r.RuleID, r.Name)
			if err := c.workforceai.DeleteRule(ctx, r.RuleID); err != nil {
				t.Logf("    delete failed: %v", err)
			}
		}
	}

	// Browse rulebases
	for _, getRB := range []func(context.Context) (*sdk.Rulebase, error){
		c.browse.GetChatsRulebase,
		c.browse.GetAccessRulebase,
		c.browse.GetSecureBrowsingRulebase,
	} {
		rb, err := getRB(ctx)
		if err != nil {
			t.Logf("  skip rulebase (get failed): %v", err)
			continue
		}
		for _, r := range rb.Rules {
			t.Logf("  deleting browse rule %s (%s)", r.RuleID, r.Name)
			if err := c.browse.DeleteRule(ctx, r.RuleID); err != nil {
				t.Logf("    delete failed: %v", err)
			}
		}
	}

	// Browse objects
	fpObjs, err := c.browse.ListFileProtectionObjects(ctx)
	if err == nil {
		for _, o := range fpObjs {
			t.Logf("  deleting file protection object %s (%s)", o.ObjectID, o.Name)
			c.browse.DeleteObject(ctx, sdk.DeleteObjectRequest{ObjectID: o.ObjectID})
		}
	}
	domObjs, err := c.browse.ListDomainsObjects(ctx)
	if err == nil {
		for _, o := range domObjs {
			t.Logf("  deleting domains object %s (%s)", o.ObjectID, o.Name)
			c.browse.DeleteObject(ctx, sdk.DeleteObjectRequest{ObjectID: o.ObjectID})
		}
	}

	// Wait for API to propagate deletes
	time.Sleep(2 * time.Second)
}

// testID returns a unique suffix for test resource names.
func testID() string {
	return fmt.Sprintf("%d", rand.Intn(99999))
}

// ---------- tests ----------

func TestIntegration(t *testing.T) {
	c := newClients(t)
	ctx := context.Background()

	// ── Phase 1: Clean slate ──
	t.Log("Phase 1: Cleaning up all existing rules and objects")
	cleanupAll(t, c)

	// ── Phase 2: WorkforceAI Chats Rule CRUD ──
	t.Run("WorkforceAI_Chats_Rule", func(t *testing.T) {
		policy := map[string]interface{}{
			"event_type": "file_upload",
			"action":     "prevent",
			"logging":    "enabled",
			"services_and_application": map[string]interface{}{
				"mode": "all",
			},
			"data_types": []map[string]interface{}{
				{"id": "cf0523c1-537e-4a4b-8bb8-084b7b9e0b45", "name": "Credit Card Number", "type": "PRE_DEFINED"},
			},
		}

		// CREATE
		rule, err := c.workforceai.CreateChatsRule(ctx, sdk.AddRuleRequest{
			Name:        "Test DLP Rule",
			Description: "Integration test",
			Order:       0,
			Policy:      policy,
			Source:      []sdk.Assignment{{AssignmentID: "entire-org", DisplayName: "Entire Organization", AssignmentType: "ASSIGNMENT_TYPE_ENTIRE_ORG"}},
		})
		if err != nil {
			t.Fatalf("create failed: %v", err)
		}
		t.Logf("created rule %s", rule.RuleID)

		// READ
		got, err := c.workforceai.GetChatsRule(ctx, rule.RuleID)
		if err != nil {
			t.Fatalf("read failed: %v", err)
		}
		assertEqual(t, "name", got.Name, "Test DLP Rule")
		assertEqual(t, "active", got.Active, true)

		// UPDATE name
		err = c.workforceai.SetRuleInfo(ctx, sdk.SetInfoRequest{
			RuleID: rule.RuleID, Version: got.Version,
			Name: "Test DLP Rule Updated", Description: "Updated",
		})
		if err != nil {
			t.Fatalf("update info failed: %v", err)
		}

		got2, err := c.workforceai.GetChatsRule(ctx, rule.RuleID)
		if err != nil {
			t.Fatalf("read after update failed: %v", err)
		}
		assertEqual(t, "updated name", got2.Name, "Test DLP Rule Updated")
		assertEqual(t, "updated desc", got2.Description, "Updated")

		// UPDATE policy (change action)
		updatedPolicy := map[string]interface{}{"action": "block"}
		err = c.workforceai.PatchChatsPolicy(ctx, sdk.PatchPolicyRequest{
			RuleID: rule.RuleID, Version: got2.Version, Policy: updatedPolicy,
		})
		if err != nil {
			t.Fatalf("patch policy failed: %v", err)
		}

		got3, err := c.workforceai.GetChatsRule(ctx, rule.RuleID)
		if err != nil {
			t.Fatalf("read after policy patch failed: %v", err)
		}
		if got3.Policy["action"] != "block" {
			t.Fatalf("expected action=block, got %v", got3.Policy["action"])
		}

		// DISABLE
		err = c.workforceai.SetRuleActive(ctx, sdk.SetActiveRequest{
			RuleID: rule.RuleID, Version: got3.Version, Active: false,
		})
		if err != nil {
			t.Fatalf("disable failed: %v", err)
		}
		got4, _ := c.workforceai.GetChatsRule(ctx, rule.RuleID)
		assertEqual(t, "active after disable", got4.Active, false)

		// DELETE
		err = c.workforceai.DeleteRule(ctx, rule.RuleID)
		if err != nil {
			t.Fatalf("delete failed: %v", err)
		}

		// VERIFY GONE
		_, err = c.workforceai.GetChatsRule(ctx, rule.RuleID)
		if err == nil {
			t.Fatal("expected error reading deleted rule, got nil")
		}
	})

	// ── Phase 3: WorkforceAI Access Rule CRUD ──
	t.Run("WorkforceAI_Access_Rule", func(t *testing.T) {
		policy := map[string]interface{}{
			"action":  "block",
			"logging": "enabled",
			"services_and_application": map[string]interface{}{
				"mode": "selected",
				"genai_application": []map[string]interface{}{
					{"id": 1, "mode": "all"},
				},
			},
		}

		rule, err := c.workforceai.CreateAccessRule(ctx, sdk.AddRuleRequest{
			Name:   "Test Access Rule",
			Order:  0,
			Policy: policy,
			Source:  []sdk.Assignment{{AssignmentID: "entire-org", DisplayName: "Entire Organization", AssignmentType: "ASSIGNMENT_TYPE_ENTIRE_ORG"}},
		})
		if err != nil {
			t.Fatalf("create failed: %v", err)
		}
		t.Logf("created rule %s", rule.RuleID)

		got, _ := c.workforceai.GetAccessRule(ctx, rule.RuleID)
		assertEqual(t, "name", got.Name, "Test Access Rule")

		// Update and verify
		c.workforceai.SetRuleInfo(ctx, sdk.SetInfoRequest{
			RuleID: rule.RuleID, Version: got.Version,
			Name: "Test Access Rule v2", Description: "v2",
		})
		got2, _ := c.workforceai.GetAccessRule(ctx, rule.RuleID)
		assertEqual(t, "updated name", got2.Name, "Test Access Rule v2")

		// Delete
		c.workforceai.DeleteRule(ctx, rule.RuleID)
		_, err = c.workforceai.GetAccessRule(ctx, rule.RuleID)
		if err == nil {
			t.Fatal("expected error after delete")
		}
	})

	// ── Phase 4: WorkforceAI Agents Rule CRUD ──
	t.Run("WorkforceAI_Agents_Rule", func(t *testing.T) {
		policy := map[string]interface{}{
			"action":  "allow",
			"logging": "enabled",
			"clients": map[string]interface{}{"mode": "all"},
			"servers": map[string]interface{}{"mcp_servers_mode": "all"},
			"tooling": map[string]interface{}{"match_mode": "all"},
		}

		rule, err := c.workforceai.CreateAgentsRule(ctx, sdk.AddRuleRequest{
			Name:   "Test Agents Rule",
			Order:  0,
			Policy: policy,
			Source:  []sdk.Assignment{{AssignmentID: "entire-org", DisplayName: "Entire Organization", AssignmentType: "ASSIGNMENT_TYPE_ENTIRE_ORG"}},
		})
		if err != nil {
			t.Fatalf("create failed: %v", err)
		}
		t.Logf("created rule %s", rule.RuleID)

		// Verify read-back
		got, _ := c.workforceai.GetAgentsRule(ctx, rule.RuleID)
		assertEqual(t, "name", got.Name, "Test Agents Rule")

		// Delete
		c.workforceai.DeleteRule(ctx, rule.RuleID)
	})

	// ── Phase 5: Browse Domains Object CRUD ──
	t.Run("Browse_Domains_Object", func(t *testing.T) {
		name := "Test Domains " + testID()
		created, err := c.browse.CreateDomainsObject(ctx, sdk.DomainsObject{
			Name:        name,
			Description: "Integration test",
			ObjectData:  json.RawMessage(`{"domains":[{"domain":"test.example.com","match_mode":"exact_host"}]}`),
		})
		if err != nil {
			t.Fatalf("create failed: %v", err)
		}
		t.Logf("created domains object %s", created.ObjectID)

		// READ
		got, err := c.browse.GetDomainsObject(ctx, created.ObjectID)
		if err != nil {
			t.Fatalf("read failed: %v", err)
		}
		assertEqual(t, "name", got.Name, name)

		// UPDATE
		updatedName := name + " Updated"
		_, err = c.browse.UpdateDomainsObject(ctx, created.ObjectID, sdk.DomainsObject{
			Name:        updatedName,
			Description: "Updated",
			ObjectData:  json.RawMessage(`{"domains":[{"domain":"updated.example.com","match_mode":"any_subdomain"}]}`),
		})
		if err != nil {
			t.Fatalf("update failed: %v", err)
		}

		got2, _ := c.browse.GetDomainsObject(ctx, created.ObjectID)
		assertEqual(t, "updated name", got2.Name, updatedName)

		// Verify object_data round-trip
		var domains struct {
			Domains []struct {
				Domain string `json:"domain"`
			} `json:"domains"`
		}
		json.Unmarshal(got2.ObjectData, &domains)
		if len(domains.Domains) != 1 || domains.Domains[0].Domain != "updated.example.com" {
			t.Fatalf("unexpected domains after update: %s", string(got2.ObjectData))
		}

		// DELETE
		err = c.browse.DeleteObject(ctx, sdk.DeleteObjectRequest{ObjectID: created.ObjectID})
		if err != nil {
			t.Fatalf("delete failed: %v", err)
		}
		t.Log("domains object deleted successfully")
	})

	// ── Phase 6: Browse File Protection Object CRUD ──
	t.Run("Browse_File_Protection_Object", func(t *testing.T) {
		name := "Test FP " + testID()
		created, err := c.browse.CreateFileProtectionObject(ctx, sdk.FileProtectionObject{
			Name:        name,
			Description: "Integration test",
			ObjectData:  json.RawMessage(`{"file_download_action_unsupported":"Block","file_upload_action_unsupported":"Block","file_download_size_limit_mb":50,"file_upload_size_limit_mb":50}`),
		})
		if err != nil {
			t.Fatalf("create failed: %v", err)
		}
		t.Logf("created file protection object %s", created.ObjectID)

		got, _ := c.browse.GetFileProtectionObject(ctx, created.ObjectID)
		assertEqual(t, "name", got.Name, name)

		// UPDATE
		updatedName := name + " Updated"
		_, err = c.browse.UpdateFileProtectionObject(ctx, created.ObjectID, sdk.FileProtectionObject{
			Name:        updatedName,
			Description: "Updated",
			ObjectData:  json.RawMessage(`{"file_download_action_unsupported":"Allow","file_upload_action_unsupported":"Block"}`),
		})
		if err != nil {
			t.Fatalf("update failed: %v", err)
		}

		got2, _ := c.browse.GetFileProtectionObject(ctx, created.ObjectID)
		assertEqual(t, "updated name", got2.Name, updatedName)

		c.browse.DeleteObject(ctx, sdk.DeleteObjectRequest{ObjectID: created.ObjectID})
	})

	// ── Phase 7: Browse Secure Browsing Rule with Objects ──
	t.Run("Browse_Secure_Browsing_Rule", func(t *testing.T) {
		// Create a domains object first
		domName := "SB Domains " + testID()
		domObj, err := c.browse.CreateDomainsObject(ctx, sdk.DomainsObject{
			Name:       domName,
			ObjectData: json.RawMessage(`{"domains":[{"domain":"secure.example.com","match_mode":"exact_host"}]}`),
		})
		if err != nil {
			t.Fatalf("create domains object failed: %v", err)
		}

		policy := map[string]interface{}{
			"password_reuse":    "prevent",
			"zero_phishing":     "detect",
			"safe_search":       "on",
			"search_reputation": "on",
		}

		rule, err := c.browse.CreateSecureBrowsingRule(ctx, sdk.AddRuleRequest{
			Name:   "Test Secure Browsing " + testID(),
			Order:  0,
			Policy: policy,
			Source:  []sdk.Assignment{{AssignmentID: "entire-org", DisplayName: "Entire Organization", AssignmentType: "ASSIGNMENT_TYPE_ENTIRE_ORG"}},
		})
		if err != nil {
			t.Fatalf("create rule failed: %v", err)
		}
		t.Logf("created secure browsing rule %s", rule.RuleID)

		// Attach objects
		err = c.browse.SetSecureBrowsingObjects(ctx, sdk.SetObjectsRequest{
			RuleID:    rule.RuleID,
			Feature:   "PROTECTED_DOMAINS",
			ObjectIDs: []string{domObj.ObjectID},
		})
		if err != nil {
			t.Fatalf("set objects failed: %v", err)
		}

		// Read and verify objects attached
		got, _ := c.browse.GetSecureBrowsingRule(ctx, rule.RuleID)
		if ids, ok := got.Objects["PROTECTED_DOMAINS"]; !ok || len(ids) != 1 {
			t.Fatalf("expected 1 PROTECTED_DOMAINS object, got %v", got.Objects)
		}

		// Update policy
		err = c.browse.PatchSecureBrowsingPolicy(ctx, sdk.PatchPolicyRequest{
			RuleID: rule.RuleID, Version: got.Version,
			Policy: map[string]interface{}{"safe_search": "off"},
		})
		if err != nil {
			t.Fatalf("patch policy failed: %v", err)
		}
		got2, _ := c.browse.GetSecureBrowsingRule(ctx, rule.RuleID)
		if got2.Policy["safe_search"] != "off" {
			t.Fatalf("expected safe_search=off, got %v", got2.Policy["safe_search"])
		}

		// Cleanup
		c.browse.DeleteRule(ctx, rule.RuleID)
		c.browse.DeleteObject(ctx, sdk.DeleteObjectRequest{ObjectID: domObj.ObjectID})
	})

	// ── Phase 8: Browse Access Rule ──
	t.Run("Browse_Access_Rule", func(t *testing.T) {
		id := testID()
		domObj, err := c.browse.CreateDomainsObject(ctx, sdk.DomainsObject{
			Name:       "Access Domains " + id,
			ObjectData: json.RawMessage(`{"domains":[{"domain":"access.example.com","match_mode":"exact_host"}]}`),
		})
		if err != nil {
			t.Fatalf("create domains object failed: %v", err)
		}
		fpObj, err := c.browse.CreateFileProtectionObject(ctx, sdk.FileProtectionObject{
			Name:       "Access FP " + id,
			ObjectData: json.RawMessage(`{"file_download_action_unsupported":"Block","file_upload_action_unsupported":"Block"}`),
		})
		if err != nil {
			t.Fatalf("create file protection object failed: %v", err)
		}

		rule, err := c.browse.CreateAccessRule(ctx, sdk.AddRuleRequest{
			Name:  "Test Access " + id,
			Order: 0,
			Policy: map[string]interface{}{
				"action": "allow", "logging": "enabled",
				"services_and_application": map[string]interface{}{"mode": "all"},
			},
			Source: []sdk.Assignment{{AssignmentID: "entire-org", DisplayName: "Entire Organization", AssignmentType: "ASSIGNMENT_TYPE_ENTIRE_ORG"}},
		})
		if err != nil {
			t.Fatalf("create failed: %v", err)
		}
		t.Logf("created access rule %s", rule.RuleID)

		// Attach objects
		c.browse.SetAccessObjects(ctx, sdk.SetObjectsRequest{
			RuleID: rule.RuleID, Feature: "DOMAIN_SERVICES", ObjectIDs: []string{domObj.ObjectID},
		})
		c.browse.SetAccessObjects(ctx, sdk.SetObjectsRequest{
			RuleID: rule.RuleID, Feature: "FILE_EMULATION", ObjectIDs: []string{fpObj.ObjectID},
		})

		got, _ := c.browse.GetAccessRule(ctx, rule.RuleID)
		t.Logf("objects after set: %v", got.Objects)

		// Cleanup
		c.browse.DeleteRule(ctx, rule.RuleID)
		c.browse.DeleteObject(ctx, sdk.DeleteObjectRequest{ObjectID: domObj.ObjectID})
		c.browse.DeleteObject(ctx, sdk.DeleteObjectRequest{ObjectID: fpObj.ObjectID})
	})

	// ── Phase 9: Browse DLP Rule ──
	t.Run("Browse_Dlp_Rule", func(t *testing.T) {
		id := testID()
		policy := map[string]interface{}{
			"event_type": "file_upload",
			"action":     "block",
			"logging":    "enabled",
			"services_and_application": map[string]interface{}{"mode": "all"},
		}

		rule, err := c.browse.CreateChatsRule(ctx, sdk.AddRuleRequest{
			Name:   "Test Browse DLP " + id,
			Order:  0,
			Policy: policy,
			Source:  []sdk.Assignment{{AssignmentID: "entire-org", DisplayName: "Entire Organization", AssignmentType: "ASSIGNMENT_TYPE_ENTIRE_ORG"}},
		})
		if err != nil {
			t.Fatalf("create failed: %v", err)
		}
		t.Logf("created browse dlp rule %s", rule.RuleID)

		got, _ := c.browse.GetChatsRule(ctx, rule.RuleID)
		assertEqual(t, "name", got.Name, "Test Browse DLP "+id)

		c.browse.DeleteRule(ctx, rule.RuleID)
	})

	// ── Phase 11: Idempotency check ──
	t.Run("Idempotency", func(t *testing.T) {
		policy := map[string]interface{}{
			"password_reuse":    "prevent",
			"zero_phishing":     "prevent",
			"safe_search":       "on",
			"search_reputation": "on",
		}

		rule, err := c.browse.CreateSecureBrowsingRule(ctx, sdk.AddRuleRequest{
			Name:   "Idempotency " + testID(),
			Order:  0,
			Policy: policy,
			Source:  []sdk.Assignment{{AssignmentID: "entire-org", DisplayName: "Entire Organization", AssignmentType: "ASSIGNMENT_TYPE_ENTIRE_ORG"}},
		})
		if err != nil {
			t.Fatalf("create failed: %v", err)
		}

		// Read twice and compare
		got1, _ := c.browse.GetSecureBrowsingRule(ctx, rule.RuleID)
		got2, _ := c.browse.GetSecureBrowsingRule(ctx, rule.RuleID)

		j1, _ := json.Marshal(got1.Policy)
		j2, _ := json.Marshal(got2.Policy)
		if string(j1) != string(j2) {
			t.Fatalf("idempotency failure: read1=%s read2=%s", j1, j2)
		}

		// Verify version didn't change (no mutation)
		if got1.Version != got2.Version {
			t.Fatalf("version changed between reads: %d -> %d", got1.Version, got2.Version)
		}

		c.browse.DeleteRule(ctx, rule.RuleID)
	})

	// ── Phase 12: SetSource — change source after create ──
	t.Run("SetSource", func(t *testing.T) {
		id := testID()
		rule, err := c.workforceai.CreateChatsRule(ctx, sdk.AddRuleRequest{
			Name:  "Source Test " + id,
			Order: 0,
			Policy: map[string]interface{}{
				"event_type": "file_upload", "action": "block", "logging": "enabled",
				"services_and_application": map[string]interface{}{"mode": "all"},
			},
			Source: []sdk.Assignment{{AssignmentID: "entire-org", DisplayName: "Entire Organization", AssignmentType: "ASSIGNMENT_TYPE_ENTIRE_ORG"}},
		})
		if err != nil {
			t.Fatalf("create failed: %v", err)
		}

		// Verify initial source
		got, _ := c.workforceai.GetChatsRule(ctx, rule.RuleID)
		if len(got.Source) == 0 {
			t.Fatal("expected non-empty source after create")
		}
		assertEqual(t, "initial source type", got.Source[0].AssignmentType, "ASSIGNMENT_TYPE_ENTIRE_ORG")

		// Change source to empty (clear assignments)
		err = c.workforceai.SetRuleSource(ctx, sdk.SetSourceRequest{
			RuleID: rule.RuleID,
			Source: []sdk.Assignment{},
		})
		if err != nil {
			t.Fatalf("set empty source failed: %v", err)
		}

		got2, _ := c.workforceai.GetChatsRule(ctx, rule.RuleID)
		if len(got2.Source) != 0 {
			t.Fatalf("expected empty source, got %d entries", len(got2.Source))
		}

		// Set source back to entire org
		err = c.workforceai.SetRuleSource(ctx, sdk.SetSourceRequest{
			RuleID: rule.RuleID,
			Source:  []sdk.Assignment{{AssignmentID: "entire-org", DisplayName: "Entire Organization", AssignmentType: "ASSIGNMENT_TYPE_ENTIRE_ORG"}},
		})
		if err != nil {
			t.Fatalf("set source back failed: %v", err)
		}

		got3, _ := c.workforceai.GetChatsRule(ctx, rule.RuleID)
		if len(got3.Source) != 1 {
			t.Fatalf("expected 1 source, got %d", len(got3.Source))
		}
		assertEqual(t, "restored source type", got3.Source[0].AssignmentType, "ASSIGNMENT_TYPE_ENTIRE_ORG")

		c.workforceai.DeleteRule(ctx, rule.RuleID)
	})

	// ── Phase 13: SetActive — enable/disable toggle ──
	t.Run("SetActive_Toggle", func(t *testing.T) {
		id := testID()
		rule, err := c.browse.CreateSecureBrowsingRule(ctx, sdk.AddRuleRequest{
			Name:   "Active Toggle " + id,
			Order:  0,
			Policy: map[string]interface{}{"password_reuse": "off", "zero_phishing": "off", "safe_search": "off", "search_reputation": "off"},
			Source: []sdk.Assignment{{AssignmentID: "entire-org", DisplayName: "Entire Organization", AssignmentType: "ASSIGNMENT_TYPE_ENTIRE_ORG"}},
		})
		if err != nil {
			t.Fatalf("create failed: %v", err)
		}

		// Should be active by default
		got, _ := c.browse.GetSecureBrowsingRule(ctx, rule.RuleID)
		assertEqual(t, "default active", got.Active, true)

		// Disable
		err = c.browse.SetRuleActive(ctx, sdk.SetActiveRequest{
			RuleID: rule.RuleID, Version: got.Version, Active: false,
		})
		if err != nil {
			t.Fatalf("disable failed: %v", err)
		}
		got2, _ := c.browse.GetSecureBrowsingRule(ctx, rule.RuleID)
		assertEqual(t, "disabled", got2.Active, false)

		// Re-enable
		err = c.browse.SetRuleActive(ctx, sdk.SetActiveRequest{
			RuleID: rule.RuleID, Version: got2.Version, Active: true,
		})
		if err != nil {
			t.Fatalf("re-enable failed: %v", err)
		}
		got3, _ := c.browse.GetSecureBrowsingRule(ctx, rule.RuleID)
		assertEqual(t, "re-enabled", got3.Active, true)

		c.browse.DeleteRule(ctx, rule.RuleID)
	})

	// ── Phase 14: SetPolicy (full replace) vs PatchPolicy ──
	t.Run("SetPolicy_FullReplace", func(t *testing.T) {
		id := testID()
		rule, err := c.workforceai.CreateAccessRule(ctx, sdk.AddRuleRequest{
			Name:  "SetPolicy Test " + id,
			Order: 0,
			Policy: map[string]interface{}{
				"action": "block", "logging": "enabled",
				"services_and_application": map[string]interface{}{"mode": "all"},
			},
			Source: []sdk.Assignment{{AssignmentID: "entire-org", DisplayName: "Entire Organization", AssignmentType: "ASSIGNMENT_TYPE_ENTIRE_ORG"}},
		})
		if err != nil {
			t.Fatalf("create failed: %v", err)
		}

		got, _ := c.workforceai.GetAccessRule(ctx, rule.RuleID)
		assertEqual(t, "initial action", fmt.Sprintf("%v", got.Policy["action"]), "block")

		// Full replace policy
		err = c.workforceai.SetAccessPolicy(ctx, sdk.SetPolicyRequest{
			RuleID:  rule.RuleID,
			Version: got.Version,
			Policy: map[string]interface{}{
				"action": "allow", "logging": "disabled",
				"services_and_application": map[string]interface{}{"mode": "all"},
			},
		})
		if err != nil {
			t.Fatalf("set policy failed: %v", err)
		}

		got2, _ := c.workforceai.GetAccessRule(ctx, rule.RuleID)
		assertEqual(t, "replaced action", fmt.Sprintf("%v", got2.Policy["action"]), "allow")
		assertEqual(t, "replaced logging", fmt.Sprintf("%v", got2.Policy["logging"]), "disabled")

		c.workforceai.DeleteRule(ctx, rule.RuleID)
	})

	// ── Phase 15: Reorder — create 2 rules and swap order ──
	t.Run("Reorder", func(t *testing.T) {
		id := testID()
		rule1, err := c.browse.CreateSecureBrowsingRule(ctx, sdk.AddRuleRequest{
			Name:   "Reorder First " + id,
			Order:  0,
			Policy: map[string]interface{}{"password_reuse": "off"},
			Source: []sdk.Assignment{{AssignmentID: "entire-org", DisplayName: "Entire Organization", AssignmentType: "ASSIGNMENT_TYPE_ENTIRE_ORG"}},
		})
		if err != nil {
			t.Fatalf("create rule1 failed: %v", err)
		}

		rule2, err := c.browse.CreateSecureBrowsingRule(ctx, sdk.AddRuleRequest{
			Name:   "Reorder Second " + id,
			Order:  1,
			Policy: map[string]interface{}{"password_reuse": "off"},
			Source: []sdk.Assignment{{AssignmentID: "entire-org", DisplayName: "Entire Organization", AssignmentType: "ASSIGNMENT_TYPE_ENTIRE_ORG"}},
		})
		if err != nil {
			t.Fatalf("create rule2 failed: %v", err)
		}

		// Verify initial order
		got1, _ := c.browse.GetSecureBrowsingRule(ctx, rule1.RuleID)
		got2, _ := c.browse.GetSecureBrowsingRule(ctx, rule2.RuleID)
		t.Logf("before reorder: rule1 order=%d, rule2 order=%d", got1.Order, got2.Order)

		// Move rule2 to position 0 (before rule1)
		err = c.browse.ReorderRule(ctx, sdk.ReorderRequest{
			RuleID:   rule2.RuleID,
			Version:  got2.Version,
			NewOrder: 0,
		})
		if err != nil {
			t.Fatalf("reorder failed: %v", err)
		}

		// Verify new order
		got1After, _ := c.browse.GetSecureBrowsingRule(ctx, rule1.RuleID)
		got2After, _ := c.browse.GetSecureBrowsingRule(ctx, rule2.RuleID)
		t.Logf("after reorder: rule1 order=%d, rule2 order=%d", got1After.Order, got2After.Order)

		if got2After.Order >= got1After.Order {
			t.Fatalf("expected rule2 order < rule1 order after reorder, got rule2=%d rule1=%d", got2After.Order, got1After.Order)
		}

		c.browse.DeleteRule(ctx, rule1.RuleID)
		c.browse.DeleteRule(ctx, rule2.RuleID)
	})

	// ── Phase 16: GetRulebase — list all rules ──
	t.Run("GetRulebase", func(t *testing.T) {
		id := testID()
		r1, _ := c.workforceai.CreateChatsRule(ctx, sdk.AddRuleRequest{
			Name:  "Rulebase List A " + id,
			Order: 0,
			Policy: map[string]interface{}{
				"event_type": "file_upload", "action": "block", "logging": "enabled",
				"services_and_application": map[string]interface{}{"mode": "all"},
			},
			Source: []sdk.Assignment{{AssignmentID: "entire-org", DisplayName: "Entire Organization", AssignmentType: "ASSIGNMENT_TYPE_ENTIRE_ORG"}},
		})
		r2, _ := c.workforceai.CreateChatsRule(ctx, sdk.AddRuleRequest{
			Name:  "Rulebase List B " + id,
			Order: 1,
			Policy: map[string]interface{}{
				"event_type": "paste", "action": "allow", "logging": "enabled",
				"services_and_application": map[string]interface{}{"mode": "all"},
			},
			Source: []sdk.Assignment{{AssignmentID: "entire-org", DisplayName: "Entire Organization", AssignmentType: "ASSIGNMENT_TYPE_ENTIRE_ORG"}},
		})

		rb, err := c.workforceai.GetChatsRulebase(ctx)
		if err != nil {
			t.Fatalf("get rulebase failed: %v", err)
		}
		t.Logf("rulebase has %d rules, version=%d", len(rb.Rules), rb.RulebaseVersion)

		if len(rb.Rules) < 2 {
			t.Fatalf("expected at least 2 rules in rulebase, got %d", len(rb.Rules))
		}

		// Verify our rules are in the list
		found := 0
		for _, r := range rb.Rules {
			if r.RuleID == r1.RuleID || r.RuleID == r2.RuleID {
				found++
			}
		}
		if found != 2 {
			t.Fatalf("expected to find both rules in rulebase, found %d", found)
		}

		c.workforceai.DeleteRule(ctx, r1.RuleID)
		c.workforceai.DeleteRule(ctx, r2.RuleID)
	})

	// ── Phase 17: Browse DLP Objects — attach DOMAIN_SERVICES to chats rule ──
	t.Run("Browse_Dlp_Objects", func(t *testing.T) {
		id := testID()
		domObj, err := c.browse.CreateDomainsObject(ctx, sdk.DomainsObject{
			Name:       "Chats Obj Domains " + id,
			ObjectData: json.RawMessage(`{"domains":[{"domain":"chats-obj.example.com","match_mode":"exact_host"}]}`),
		})
		if err != nil {
			t.Fatalf("create domains object failed: %v", err)
		}

		rule, err := c.browse.CreateChatsRule(ctx, sdk.AddRuleRequest{
			Name:  "Chats With Obj " + id,
			Order: 0,
			Policy: map[string]interface{}{
				"event_type": "file_upload", "action": "block", "logging": "enabled",
				"services_and_application": map[string]interface{}{"mode": "all"},
			},
			Source: []sdk.Assignment{{AssignmentID: "entire-org", DisplayName: "Entire Organization", AssignmentType: "ASSIGNMENT_TYPE_ENTIRE_ORG"}},
		})
		if err != nil {
			t.Fatalf("create chats rule failed: %v", err)
		}

		// Attach DOMAIN_SERVICES
		err = c.browse.SetChatsObjects(ctx, sdk.SetObjectsRequest{
			RuleID:    rule.RuleID,
			Feature:   "DOMAIN_SERVICES",
			ObjectIDs: []string{domObj.ObjectID},
		})
		if err != nil {
			t.Fatalf("set chats objects failed: %v", err)
		}

		got, _ := c.browse.GetChatsRule(ctx, rule.RuleID)
		if len(got.Objects["DOMAIN_SERVICES"]) != 1 {
			t.Fatalf("expected 1 DOMAIN_SERVICES, got %v", got.Objects)
		}

		// Remove objects (set empty)
		err = c.browse.SetChatsObjects(ctx, sdk.SetObjectsRequest{
			RuleID:    rule.RuleID,
			Feature:   "DOMAIN_SERVICES",
			ObjectIDs: []string{},
		})
		if err != nil {
			t.Fatalf("clear chats objects failed: %v", err)
		}

		got2, _ := c.browse.GetChatsRule(ctx, rule.RuleID)
		if len(got2.Objects["DOMAIN_SERVICES"]) != 0 {
			t.Fatalf("expected 0 DOMAIN_SERVICES after clear, got %v", got2.Objects)
		}

		c.browse.DeleteRule(ctx, rule.RuleID)
		c.browse.DeleteObject(ctx, sdk.DeleteObjectRequest{ObjectID: domObj.ObjectID})
	})

	// ── Phase 18: Version increment — verify version bumps across mutations ──
	t.Run("Version_Tracking", func(t *testing.T) {
		id := testID()
		rule, err := c.browse.CreateSecureBrowsingRule(ctx, sdk.AddRuleRequest{
			Name:   "Version Track " + id,
			Order:  0,
			Policy: map[string]interface{}{"password_reuse": "off"},
			Source: []sdk.Assignment{{AssignmentID: "entire-org", DisplayName: "Entire Organization", AssignmentType: "ASSIGNMENT_TYPE_ENTIRE_ORG"}},
		})
		if err != nil {
			t.Fatalf("create failed: %v", err)
		}

		v0, _ := c.browse.GetSecureBrowsingRule(ctx, rule.RuleID)
		t.Logf("initial version: %d", v0.Version)

		// Mutation 1: SetRuleInfo
		err = c.browse.SetRuleInfo(ctx, sdk.SetInfoRequest{
			RuleID: rule.RuleID, Version: v0.Version,
			Name: "Version Track Updated " + id, Description: "v1",
		})
		if err != nil {
			t.Fatalf("SetRuleInfo failed: %v", err)
		}
		v1, _ := c.browse.GetSecureBrowsingRule(ctx, rule.RuleID)
		t.Logf("after SetRuleInfo: version=%d, name=%s", v1.Version, v1.Name)
		assertEqual(t, "name updated", v1.Name, "Version Track Updated "+id)

		// Mutation 2: PatchPolicy (use latest version from read)
		err = c.browse.PatchSecureBrowsingPolicy(ctx, sdk.PatchPolicyRequest{
			RuleID: rule.RuleID, Version: v1.Version,
			Policy: map[string]interface{}{"safe_search": "on"},
		})
		if err != nil {
			t.Fatalf("PatchPolicy failed: %v", err)
		}
		v2, _ := c.browse.GetSecureBrowsingRule(ctx, rule.RuleID)
		t.Logf("after PatchPolicy: version=%d", v2.Version)

		// Mutation 3: SetRuleActive
		err = c.browse.SetRuleActive(ctx, sdk.SetActiveRequest{
			RuleID: rule.RuleID, Version: v2.Version, Active: false,
		})
		if err != nil {
			t.Fatalf("SetRuleActive failed: %v", err)
		}
		v3, _ := c.browse.GetSecureBrowsingRule(ctx, rule.RuleID)
		t.Logf("after SetRuleActive: version=%d, active=%v", v3.Version, v3.Active)
		assertEqual(t, "disabled", v3.Active, false)

		// Verify overall version increased from initial
		if v3.Version <= v0.Version {
			t.Fatalf("version did not increase across mutations: initial=%d final=%d", v0.Version, v3.Version)
		}

		c.browse.DeleteRule(ctx, rule.RuleID)
	})

	// ── Phase 19: Cross-product isolation — browse rule must not appear in workforce_ai rulebase and vice versa ──
	t.Run("Cross_Product_Isolation", func(t *testing.T) {
		id := testID()

		// Create a browse access rule
		browseRule, err := c.browse.CreateAccessRule(ctx, sdk.AddRuleRequest{
			Name:  "Cross Iso Browse " + id,
			Order: 0,
			Policy: map[string]interface{}{
				"action": "allow", "logging": "enabled",
				"services_and_application": map[string]interface{}{"mode": "all"},
			},
			Source: []sdk.Assignment{{AssignmentID: "entire-org", DisplayName: "Entire Organization", AssignmentType: "ASSIGNMENT_TYPE_ENTIRE_ORG"}},
		})
		if err != nil {
			t.Fatalf("create browse access rule failed: %v", err)
		}
		t.Logf("created browse access rule %s", browseRule.RuleID)

		// Create a workforce_ai access rule
		workforceAIRule, err := c.workforceai.CreateAccessRule(ctx, sdk.AddRuleRequest{
			Name:  "Cross Iso WorkforceAI " + id,
			Order: 0,
			Policy: map[string]interface{}{
				"action": "block", "logging": "enabled",
				"services_and_application": map[string]interface{}{
					"mode": "selected",
					"genai_application": []map[string]interface{}{
						{"id": 1, "mode": "all"},
					},
				},
			},
			Source: []sdk.Assignment{{AssignmentID: "entire-org", DisplayName: "Entire Organization", AssignmentType: "ASSIGNMENT_TYPE_ENTIRE_ORG"}},
		})
		if err != nil {
			t.Fatalf("create workforce_ai access rule failed: %v", err)
		}
		t.Logf("created workforce_ai access rule %s", workforceAIRule.RuleID)

		// Verify browse rule does NOT appear in workforce_ai access rulebase
		workforceAIRB, err := c.workforceai.GetAccessRulebase(ctx)
		if err != nil {
			t.Fatalf("get workforce_ai access rulebase failed: %v", err)
		}
		for _, r := range workforceAIRB.Rules {
			if r.RuleID == browseRule.RuleID {
				t.Fatalf("browse rule %s found in workforce_ai access rulebase — cross-product leak!", browseRule.RuleID)
			}
		}
		t.Log("confirmed: browse rule not in workforce_ai access rulebase")

		// Verify workforce_ai rule does NOT appear in browse access rulebase
		browseRB, err := c.browse.GetAccessRulebase(ctx)
		if err != nil {
			t.Fatalf("get browse access rulebase failed: %v", err)
		}
		for _, r := range browseRB.Rules {
			if r.RuleID == workforceAIRule.RuleID {
				t.Fatalf("workforce_ai rule %s found in browse access rulebase — cross-product leak!", workforceAIRule.RuleID)
			}
		}
		t.Log("confirmed: workforce_ai rule not in browse access rulebase")

		// Also verify each rule IS in its own rulebase
		foundWorkforceAI := false
		for _, r := range workforceAIRB.Rules {
			if r.RuleID == workforceAIRule.RuleID {
				foundWorkforceAI = true
				break
			}
		}
		if !foundWorkforceAI {
			t.Fatalf("workforce_ai rule %s not found in its own rulebase", workforceAIRule.RuleID)
		}

		foundBrowse := false
		for _, r := range browseRB.Rules {
			if r.RuleID == browseRule.RuleID {
				foundBrowse = true
				break
			}
		}
		if !foundBrowse {
			t.Fatalf("browse rule %s not found in its own rulebase", browseRule.RuleID)
		}
		t.Log("confirmed: each rule appears in its own rulebase only")

		// Cleanup
		c.browse.DeleteRule(ctx, browseRule.RuleID)
		c.workforceai.DeleteRule(ctx, workforceAIRule.RuleID)
	})

	// ── Phase 20: Final cleanup ──
	t.Log("Phase 20: Final cleanup")
	cleanupAll(t, c)
}

// assertEqual is a simple test helper.
func assertEqual[T comparable](t *testing.T, field string, got, want T) {
	t.Helper()
	if got != want {
		t.Fatalf("%s: got %v, want %v", field, got, want)
	}
}

// formatJSON pretty-prints JSON for debugging.
func formatJSON(data interface{}) string {
	b, _ := json.MarshalIndent(data, "", "  ")
	return string(b)
}

// ptr returns a pointer to the given value.
func ptr[T any](v T) *T { return &v }

// suppress unused warnings
var _ = formatJSON
var _ = ptr[string]
var _ = fmt.Sprintf
