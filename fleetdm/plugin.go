package fleetdm

import (
	"context"

	"github.com/turbot/steampipe-plugin-sdk/v6/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v6/plugin/transform"
)

// Plugin returns the FleetDM plugin.
func Plugin(ctx context.Context) *plugin.Plugin {
	p := &plugin.Plugin{
		Name: "steampipe-plugin-fleetdm",
		ConnectionConfigSchema: &plugin.ConnectionConfigSchema{
			NewInstance: ConfigInstance,
			Schema:      ConfigSchema,
		},
		// Note: deliberately no NullIfZero() here — it would turn legitimate
		// false/0 values (e.g. pack.disabled, label.host_count) into SQL NULL.
		DefaultTransform: transform.FromGo(),
		// Retry rate limits (429) and transient server errors (5xx) with
		// exponential backoff; treat 404s as "no rows" rather than an error.
		DefaultRetryConfig: &plugin.RetryConfig{
			ShouldRetryErrorFunc: shouldRetryError,
			MaxAttempts:          5,
			BackoffAlgorithm:     "Exponential",
			RetryInterval:        500,    // ms
			CappedDuration:       30000,  // ms
			MaxDuration:          120000, // ms
		},
		DefaultIgnoreConfig: &plugin.IgnoreConfig{
			ShouldIgnoreErrorFunc: shouldIgnoreError,
		},
		TableMap: map[string]*plugin.Table{
			"fleetdm_ab_token":                       tableFleetdmABToken(ctx),
			"fleetdm_activity":                       tableFleetdmActivity(ctx),
			"fleetdm_android_enterprise":             tableFleetdmAndroidEnterprise(ctx),
			"fleetdm_app_store_app":                  tableFleetdmAppStoreApp(ctx),
			"fleetdm_apple_mdm":                      tableFleetdmAppleMDM(ctx),
			"fleetdm_bootstrap_package":              tableFleetdmBootstrapPackage(ctx),
			"fleetdm_carve":                          tableFleetdmCarve(ctx),
			"fleetdm_certificate_authority":          tableFleetdmCertificateAuthority(ctx),
			"fleetdm_certificate_template":           tableFleetdmCertificateTemplate(ctx),
			"fleetdm_config":                         tableFleetdmConfig(ctx),
			"fleetdm_configuration_profile":          tableFleetdmConfigurationProfile(ctx),
			"fleetdm_configuration_profile_summary":  tableFleetdmConfigurationProfileSummary(ctx),
			"fleetdm_custom_host_vital":              tableFleetdmCustomHostVital(ctx),
			"fleetdm_custom_variable":                tableFleetdmCustomVariable(ctx),
			"fleetdm_disk_encryption_summary":        tableFleetdmDiskEncryptionSummary(ctx),
			"fleetdm_enroll_secret":                  tableFleetdmEnrollSecret(ctx),
			"fleetdm_fleet_maintained_app":           tableFleetdmFleetMaintainedApp(ctx),
			"fleetdm_host":                           tableFleetdmHost(ctx),
			"fleetdm_host_activity":                  tableFleetdmHostActivity(ctx),
			"fleetdm_host_certificate":               tableFleetdmHostCertificate(ctx),
			"fleetdm_host_configuration_profile":     tableFleetdmHostConfigurationProfile(ctx),
			"fleetdm_host_detail":                    tableFleetdmHostDetail(ctx),
			"fleetdm_host_macadmins":                 tableFleetdmHostMacadmins(ctx),
			"fleetdm_host_script":                    tableFleetdmHostScript(ctx),
			"fleetdm_host_software":                  tableFleetdmHostSoftware(ctx),
			"fleetdm_host_summary":                   tableFleetdmHostSummary(ctx),
			"fleetdm_host_upcoming_activity":         tableFleetdmHostUpcomingActivity(ctx),
			"fleetdm_invite":                         tableFleetdmInvite(ctx),
			"fleetdm_label":                          tableFleetdmLabel(ctx),
			"fleetdm_label_host":                     tableFleetdmLabelHost(ctx),
			"fleetdm_mdm_command":                    tableFleetdmMDMCommand(ctx),
			"fleetdm_mdm_command_result":             tableFleetdmMDMCommandResult(ctx),
			"fleetdm_mdm_summary":                    tableFleetdmMDMSummary(ctx),
			"fleetdm_os_version":                     tableFleetdmOSVersion(ctx),
			"fleetdm_pack":                           tableFleetdmPack(ctx),
			"fleetdm_policy":                         tableFleetdmPolicy(ctx),
			"fleetdm_query":                          tableFleetdmQuery(ctx),
			"fleetdm_query_report":                   tableFleetdmQueryReport(ctx),
			"fleetdm_rest_api_endpoint":              tableFleetdmRESTAPIEndpoint(ctx),
			"fleetdm_scim_status":                    tableFleetdmScimStatus(ctx),
			"fleetdm_script":                         tableFleetdmScript(ctx),
			"fleetdm_script_result":                  tableFleetdmScriptResult(ctx),
			"fleetdm_setup_experience_software":      tableFleetdmSetupExperienceSoftware(ctx),
			"fleetdm_software_install_result":        tableFleetdmSoftwareInstallResult(ctx),
			"fleetdm_software_self_service_category": tableFleetdmSoftwareSelfServiceCategory(ctx),
			"fleetdm_software_title":                 tableFleetdmSoftwareTitle(ctx),
			"fleetdm_software_version":               tableFleetdmSoftwareVersion(ctx),
			"fleetdm_team":                           tableFleetdmTeam(ctx),
			"fleetdm_user":                           tableFleetdmUser(ctx),
			"fleetdm_version":                        tableFleetdmVersion(ctx),
			"fleetdm_vpp_token":                      tableFleetdmVPPToken(ctx),
			"fleetdm_vulnerability":                  tableFleetdmVulnerability(ctx),
		},
	}
	return p
}
