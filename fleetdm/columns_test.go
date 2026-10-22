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
		"fleetdm_ab_token":                       ABToken{},
		"fleetdm_activity":                       Activity{},
		"fleetdm_android_enterprise":             AndroidEnterprise{},
		"fleetdm_app_store_app":                  AppStoreAppWithTeam{},
		"fleetdm_apple_mdm":                      AppleMDM{},
		"fleetdm_bootstrap_package":              BootstrapPackage{},
		"fleetdm_carve":                          Carve{},
		"fleetdm_certificate_authority":          CertificateAuthority{},
		"fleetdm_certificate_template":           CertificateTemplate{},
		"fleetdm_config":                         AppConfig{},
		"fleetdm_configuration_profile":          ConfigurationProfile{},
		"fleetdm_configuration_profile_summary":  ConfigurationProfileSummary{},
		"fleetdm_custom_host_vital":              CustomHostVital{},
		"fleetdm_custom_variable":                CustomVariable{},
		"fleetdm_disk_encryption_summary":        DiskEncryptionStatus{},
		"fleetdm_enroll_secret":                  EnrollSecret{},
		"fleetdm_fleet_maintained_app":           FleetMaintainedApp{},
		"fleetdm_host":                           Host{},
		"fleetdm_host_activity":                  HostActivity{},
		"fleetdm_host_certificate":               HostCertificate{},
		"fleetdm_host_configuration_profile":     HostConfigurationProfile{},
		"fleetdm_host_detail":                    Host{},
		"fleetdm_host_macadmins":                 HostMacadmins{},
		"fleetdm_host_script":                    HostScript{},
		"fleetdm_host_software":                  HostSoftwareItem{},
		"fleetdm_host_summary":                   HostSummary{},
		"fleetdm_host_upcoming_activity":         HostActivity{},
		"fleetdm_invite":                         Invite{},
		"fleetdm_label":                          Label{},
		"fleetdm_label_host":                     LabelHost{},
		"fleetdm_mdm_command":                    MDMCommand{},
		"fleetdm_mdm_command_result":             MDMCommandResult{},
		"fleetdm_mdm_summary":                    MDMEnrollmentSummary{},
		"fleetdm_os_version":                     OSVersion{},
		"fleetdm_pack":                           Pack{},
		"fleetdm_policy":                         Policy{},
		"fleetdm_query":                          QuerySaved{},
		"fleetdm_query_report":                   QueryReportRow{},
		"fleetdm_rest_api_endpoint":              RESTAPIEndpoint{},
		"fleetdm_scim_status":                    ScimStatus{},
		"fleetdm_script":                         Script{},
		"fleetdm_script_result":                  ScriptResult{},
		"fleetdm_setup_experience_software":      SetupExperienceSoftware{},
		"fleetdm_software_install_result":        SoftwareInstallResult{},
		"fleetdm_software_self_service_category": SelfServiceCategory{},
		"fleetdm_software_title":                 SoftwareTitle{},
		"fleetdm_software_version":               Software{},
		"fleetdm_team":                           Team{},
		"fleetdm_user":                           User{},
		"fleetdm_version":                        VersionInfo{},
		"fleetdm_vpp_token":                      VPPToken{},
		"fleetdm_vulnerability":                  Vulnerability{},
	}
	hydrateRow := map[uintptr]any{
		reflect.ValueOf(getHostDetails).Pointer():          HostDetail{},
		reflect.ValueOf(getVulnerabilityDetails).Pointer(): VulnerabilityDetails{},
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
