connection "fleetdm" {
  plugin = "l-teles/fleetdm"

  # FleetDM server URL (e.g., "https://fleet.example.com")
  # The plugin will attempt to append /api/v1/ if it's not present.
  # Can also be set with the FLEETDM_URL environment variable.
  # server_url = "https://aiworld.cloud.fleetdm.com/"

  # FleetDM API Token
  # Generate this from your FleetDM instance (User Menu -> My account -> Get API token)
  # Can also be set with the FLEETDM_API_TOKEN environment variable.
  # api_token = "ZZFN9BBL+OldDhBzs61V1fRHg/2RkuYYq6qlLiDamDCCPL1vlFdHw=="

  # Opt in to returning live secret material (e.g. team enrollment secrets)
  # in query results. Leave unset/false unless you understand that secrets
  # will then land in Steampipe query caches, exports, and dashboards.
  # expose_secrets = false

  # Per-request timeout in seconds. Default: 30.
  # request_timeout = 30
}
