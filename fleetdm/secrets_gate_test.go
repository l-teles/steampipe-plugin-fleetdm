package fleetdm

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/turbot/steampipe-plugin-sdk/v6/plugin"
)

// gatedColumns lists every column that may carry credentials and must read
// NULL unless expose_secrets is set. Removing a gate fails this test.
var gatedColumns = map[string]plugin.HydrateFunc{
	"fleetdm_config.integrations":                                itemIfSecretsExposed,
	"fleetdm_config.webhook_settings":                            itemIfSecretsExposed,
	"fleetdm_config.logging":                                     itemIfSecretsExposed,
	"fleetdm_team.integrations":                                  itemIfSecretsExposed,
	"fleetdm_team.webhook_settings":                              itemIfSecretsExposed,
	"fleetdm_team.software":                                      itemIfSecretsExposed,
	"fleetdm_team.secrets":                                       getTeamSecrets,
	"fleetdm_mdm_command_result.payload":                         itemIfSecretsExposed,
	"fleetdm_mdm_command_result.result":                          itemIfSecretsExposed,
	"fleetdm_query_report.columns":                               itemIfSecretsExposed,
	"fleetdm_script_result.script_contents":                      itemIfSecretsExposed,
	"fleetdm_script_result.output":                               itemIfSecretsExposed,
	"fleetdm_software_install_result.output":                     itemIfSecretsExposed,
	"fleetdm_software_install_result.pre_install_query_output":   itemIfSecretsExposed,
	"fleetdm_software_install_result.post_install_script_output": itemIfSecretsExposed,
}

func funcPtr(f any) uintptr { return reflect.ValueOf(f).Pointer() }

func TestGatedColumnsUseSecretsGate(t *testing.T) {
	tables := Plugin(testCtx(t)).TableMap
	for key, want := range gatedColumns {
		found := false
		for name, table := range tables {
			for _, col := range table.Columns {
				if name+"."+col.Name != key {
					continue
				}
				found = true
				if col.Hydrate == nil || funcPtr(col.Hydrate) != funcPtr(want) {
					t.Errorf("%s is not gated by its expose_secrets hydrate", key)
				}
			}
		}
		if !found {
			t.Errorf("%s: column not found", key)
		}
	}
}

func TestEnrollSecretsRequireExposeSecrets(t *testing.T) {
	conn := &plugin.Connection{Name: "test"}
	conn.SetConfig(fleetdmConfig{})
	if _, err := listEnrollSecrets(testCtx(t), &plugin.QueryData{Connection: conn}, nil); err == nil {
		t.Error("fleetdm_enroll_secret must fail without expose_secrets")
	}
}

func TestNullURLKeys(t *testing.T) {
	pkg := map[string]any{"name": "app.pkg", "package_url": "https://user:token@example.com/app.pkg"}
	got := nullURLKeys(pkg, "package_url").(map[string]any)
	if got["package_url"] != nil || got["name"] != "app.pkg" {
		t.Errorf("nullURLKeys = %v", got)
	}
	if nullURLKeys(nil, "package_url") != nil {
		t.Error("nil must stay nil")
	}

	raw := nullURLKeysRaw(json.RawMessage(`{"gitops_mode_enabled":true,"repository_url":"https://x:y@git.example.com/r"}`), "repository_url")
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatal(err)
	}
	if m["repository_url"] != nil || m["gitops_mode_enabled"] != true {
		t.Errorf("nullURLKeysRaw = %s", raw)
	}
	if got := nullURLKeysRaw(json.RawMessage(`null`), "repository_url"); string(got) != "null" {
		t.Errorf("null raw = %s", got)
	}
}

func TestRedactInstallerURLs(t *testing.T) {
	conn := &plugin.Connection{Name: "test"}
	conn.SetConfig(fleetdmConfig{})
	d := &plugin.QueryData{Connection: conn}
	title := SoftwareTitle{
		SoftwarePackage: map[string]any{"name": "app.pkg", "package_url": "https://u:p@example.com/app.pkg"},
		Packages:        []any{map[string]any{"name": "app.pkg", "url": "https://example.com/app.pkg?token=x"}},
	}
	redactInstallerURLs(d, &title)
	if pkg := title.SoftwarePackage.(map[string]any); pkg["package_url"] != nil || pkg["name"] != "app.pkg" {
		t.Errorf("software_package = %v", pkg)
	}
	if pkg := title.Packages.([]any)[0].(map[string]any); pkg["url"] != nil || pkg["name"] != "app.pkg" {
		t.Errorf("packages[0] = %v", pkg)
	}

	exposed := &plugin.Connection{Name: "test"}
	exposed.SetConfig(fleetdmConfig{ExposeSecrets: ptr(true)})
	title.SoftwarePackage = map[string]any{"package_url": "kept"}
	redactInstallerURLs(&plugin.QueryData{Connection: exposed}, &title)
	if title.SoftwarePackage.(map[string]any)["package_url"] != "kept" {
		t.Error("expose_secrets = true must keep installer URLs")
	}
}
