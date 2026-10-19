package fleetdm

import (
	"context"
	"reflect"
	"testing"

	"github.com/iancoleman/strcase"
	"github.com/turbot/go-kit/helpers"
)

// TestDefaultTransformColumnsResolve guards against columns that silently read
// NULL: a column without an explicit Transform uses the SDK's FromGo, which
// maps snake_case to a Go field name and only upper-cases golint initialisms
// (so "sso_enabled" looks for SsoEnabled, never SSOEnabled).
func TestDefaultTransformColumnsResolve(t *testing.T) {
	t.Parallel()

	// Row type each table's list streams, plus the row type of any column
	// hydrate that returns a different struct.
	listRow := map[string]any{
		"fleetdm_activity":             Activity{},
		"fleetdm_app_store_app":        AppStoreAppWithTeam{},
		"fleetdm_carve":                Carve{},
		"fleetdm_fleet_maintained_app": FleetMaintainedApp{},
		"fleetdm_host":                 Host{},
		"fleetdm_host_detail":          Host{},
		"fleetdm_label":                Label{},
		"fleetdm_os_version":           OSVersion{},
		"fleetdm_pack":                 Pack{},
		"fleetdm_policy":               Policy{},
		"fleetdm_query":                QuerySaved{},
		"fleetdm_software_title":       SoftwareTitle{},
		"fleetdm_software_version":     Software{},
		"fleetdm_team":                 Team{},
		"fleetdm_user":                 User{},
		"fleetdm_version":              VersionInfo{},
	}
	hydrateRow := map[uintptr]any{
		reflect.ValueOf(getHostDetails).Pointer(): HostDetail{},
	}

	p := Plugin(context.Background())
	for name, table := range p.TableMap {
		row, ok := listRow[name]
		if !ok {
			t.Errorf("%s: no row type registered in this test", name)
			continue
		}
		for _, col := range table.Columns {
			if col.Transform != nil {
				continue
			}
			item := row
			if col.Hydrate != nil {
				if r, ok := hydrateRow[reflect.ValueOf(col.Hydrate).Pointer()]; ok {
					item = r
				}
			}
			field := helpers.LintName(strcase.ToCamel(col.Name))
			if _, ok := helpers.GetNestedFieldValueFromInterface(item, field); !ok {
				t.Errorf("%s.%s: default transform looks for field %q, which %T does not have; add an explicit Transform", name, col.Name, field, item)
			}
		}
	}
}
