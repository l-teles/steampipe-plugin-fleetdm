connection "fleetdm" {
  plugin = "l-teles/fleetdm" 

  # FleetDM server URL (e.g., "https://fleet.example.com")
  # The plugin will attempt to append /api/v1/ if it's not present.
  server_url = "YOUR_FLEETDM_SERVER_URL"

  # FleetDM API Token
  # Generate this from your FleetDM instance (User Menu -> Settings -> API Tokens)
  api_token = "YOUR_FLEETDM_API_TOKEN"
}
