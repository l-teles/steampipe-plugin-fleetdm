## v1.4.0 [unreleased]

_What's new?_

- Supports FleetDM up to v4.93.0, including servers before and after the teams→fleets and queries→reports rename (tested on 4.81 and 4.93, Premium and Free).
- 36 new tables, covering the rest of Fleet's read API:
  - Inventory and configuration: `fleetdm_config`, `fleetdm_enroll_secret`, `fleetdm_invite`, `fleetdm_custom_variable`, `fleetdm_custom_host_vital`, `fleetdm_certificate_authority`, `fleetdm_certificate_template`, `fleetdm_ab_token`, `fleetdm_vpp_token`, `fleetdm_apple_mdm`, `fleetdm_android_enterprise`, `fleetdm_bootstrap_package`, `fleetdm_setup_experience_software`, `fleetdm_software_self_service_category`, `fleetdm_rest_api_endpoint`, `fleetdm_scim_status`
  - Software, scripts and profiles: `fleetdm_vulnerability`, `fleetdm_script`, `fleetdm_configuration_profile`, `fleetdm_query_report`, `fleetdm_label_host`
  - Summaries: `fleetdm_host_summary`, `fleetdm_mdm_summary`, `fleetdm_disk_encryption_summary`, `fleetdm_configuration_profile_summary`
  - Per host: `fleetdm_host_software`, `fleetdm_host_script`, `fleetdm_host_configuration_profile`, `fleetdm_host_certificate`, `fleetdm_host_macadmins`, `fleetdm_host_activity`, `fleetdm_host_upcoming_activity`
  - Results: `fleetdm_mdm_command`, `fleetdm_mdm_command_result`, `fleetdm_script_result`, `fleetdm_software_install_result`
- New columns on existing tables for fields Fleet added up to 4.93 (host time zone, disk encryption and end users; policy type, conditional access and label scoping; user last login and MFA; team settings; and more).
- Lookups by ID (`where id = ...`) on 13 tables now make a single request instead of listing everything.
- `fleetdm_host` only asks Fleet for policies, users, labels, device mapping and end users when those columns are selected.
- New `max_requests_per_second` connection option to cap API requests.
- More data is gated behind `expose_secrets`: enroll secrets, webhook/integration/logging settings, MDM command payloads, and script/install outputs.

_Bug fixes_

- Filtering by `team_id` failed with HTTP 400 on Fleet 4.82 and later on 9 tables (`fleetdm_app_store_app` silently returned no rows).
- `fleetdm_host` and `fleetdm_host_detail` failed when a host reported an IPv6 link-local address or no public IP.
- Columns that were always empty or wrong: `fleetdm_query.team_id` for team reports, `fleetdm_user.sso_enabled`, `fleetdm_user.admin_forced_password_reset`, `fleetdm_label.built_in`, `fleetdm_team.users`.
- `fleetdm_team`, `fleetdm_fleet_maintained_app` and `fleetdm_app_store_app` failed on Fleet Free; they now return no rows.

_Dependencies_

- Upgraded to steampipe-plugin-sdk v6.1.0 and fixed vulnerable dependencies (golang.org/x/net, OpenTelemetry, gRPC, Go 1.26.9).

## v1.3.1 [2026-02-13]

_What's new?_

- Fix `fleetdm_software_version` table API field mapping issues by @l-teles in https://github.com/l-teles/steampipe-plugin-fleetdm/pull/25

## v1.3.0 [2026-02-10]

_What's new?_

- New OS Version, Software Title, Fleet-Maintained App, App Store App Tables & Software Version Rename by @l-teles in https://github.com/l-teles/steampipe-plugin-fleetdm/pull/23

- New tables
  - [fleetdm_software_title](https://hub.steampipe.io/plugins/l-teles/steampipe-plugin-fleetdm/tables/fleetdm_software_title)
  - [fleetdm_fleet_maintained_app](https://hub.steampipe.io/plugins/l-teles/steampipe-plugin-fleetdm/tables/fleetdm_fleet_maintained_app)
  - [fleetdm_os_version](https://hub.steampipe.io/plugins/l-teles/steampipe-plugin-fleetdm/tables/fleetdm_os_version)
  - [fleetdm_app_store_app](https://hub.steampipe.io/plugins/l-teles/steampipe-plugin-fleetdm/tables/fleetdm_app_store_app)

⚠️ Breaking Changes - read before upgrading

- The table `fleetdm_software` has been renamed to `fleetdm_software_version`. All queries, dashboards, and references must be updated to use the new table name.
- All pre-existing tables have been audited and updated to match only API-supported key columns. Some key columns were added, removed, or renamed to align with FleetDM REST API parameters. Queries using unsupported or legacy key columns may need to be updated.
- The `fleetdm_policy` table now dynamically switches between global and team policy endpoints based on the `team_id` key column. The `filter_search_query` key column only works when `team_id` is specified. Please go through the [PR description](https://github.com/l-teles/steampipe-plugin-fleetdm/pull/23) for details on all the changes made.

## v1.2.3 [2026-02-09]

_What's new?_

- Bump github.com/turbot/steampipe-plugin-sdk/v5 from 5.13.0 to 5.13.1 by @dependabot[bot] in https://github.com/l-teles/steampipe-plugin-fleetdm/pull/20
- Bump golang.org/x/crypto from 0.36.0 to 0.45.0 in the go_modules group across 1 directory by @dependabot[bot] in https://github.com/l-teles/steampipe-plugin-fleetdm/pull/21
- Bump github.com/turbot/steampipe-plugin-sdk/v5 from 5.13.1 to 5.13.2 by @dependabot[bot] in https://github.com/l-teles/steampipe-plugin-fleetdm/pull/22
- Fix: Handle empty time strings from FleetDM API across all tables by @l-teles in https://github.com/l-teles/steampipe-plugin-fleetdm/pull/24

## v1.2.2 [2025-08-28]

_What's new?_
- * Bump the go_modules group with 2 updates by @dependabot[bot] in https://github.com/l-teles/steampipe-plugin-fleetdm/pull/19

## v1.2.1 [2025-07-28]

_What's new?_
- 🐛 Remove `server_url` column from tables by @l-teles in https://github.com/l-teles/steampipe-plugin-fleetdm/pull/18

## v1.2.0 [2025-07-25]

_What's new?_
- New `fleetdm_carves` and `fleetdm_host_detail` tables by @l-teles in https://github.com/l-teles/steampipe-plugin-fleetdm/pull/16
- Bump github.com/turbot/steampipe-plugin-sdk/v5 from 5.12.0 to 5.13.0 by @dependabot[bot] in https://github.com/l-teles/steampipe-plugin-fleetdm/pull/17

## v1.1.1 [2025-07-25]

_What's new?_
- Bump github.com/turbot/steampipe-plugin-sdk/v5 from 5.11.6 to 5.11.7 by @dependabot in https://github.com/l-teles/steampipe-plugin-fleetdm/pull/11
- Update `dependabot.yml` by @l-teles in https://github.com/l-teles/steampipe-plugin-fleetdm/pull/12
- Bump github.com/turbot/steampipe-plugin-sdk/v5 from 5.11.7 to 5.12.0 by @dependabot in https://github.com/l-teles/steampipe-plugin-fleetdm/pull/13
- Add support for env variables for URL and API token by @l-teles in https://github.com/l-teles/steampipe-plugin-fleetdm/pull/14

## v1.1.0 [2025-05-25]

_What's new?_
- Added `device_mapping`, `labels`, `policies`, `users` columns to `fleetdm_host` table

## v1.0.1 [2025-05-26]

_What's new?_

- Minor fixes by @misraved

## v1.0.0 [2025-05-26]

_What's new?_

First plugin release 🥳

- New tables
  - [fleetdm_activity](https://hub.steampipe.io/plugins/l-teles/steampipe-plugin-fleetdm/tables/fleetdm_activity)
  - [fleetdm_host](https://hub.steampipe.io/plugins/l-teles/steampipe-plugin-fleetdm/tables/fleetdm_host)
  - [fleetdm_label](https://hub.steampipe.io/plugins/l-teles/steampipe-plugin-fleetdm/tables/fleetdm_label)
  - [fleetdm_pack](https://hub.steampipe.io/plugins/l-teles/steampipe-plugin-fleetdm/tables/fleetdm_pack)
  - [fleetdm_policy](https://hub.steampipe.io/plugins/l-teles/steampipe-plugin-fleetdm/tables/fleetdm_policy)
  - [fleetdm_query](https://hub.steampipe.io/plugins/l-teles/steampipe-plugin-fleetdm/tables/fleetdm_query)
  - [fleetdm_software](https://hub.steampipe.io/plugins/l-teles/steampipe-plugin-fleetdm/tables/fleetdm_software)
  - [fleetdm_team](https://hub.steampipe.io/plugins/l-teles/steampipe-plugin-fleetdm/tables/fleetdm_team)
  - [fleetdm_user](https://hub.steampipe.io/plugins/l-teles/steampipe-plugin-fleetdm/tables/fleetdm_user)
