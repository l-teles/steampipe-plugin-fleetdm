---
organization: FleetDM
category: ["security", "device-management"]
icon_url: "/images/plugins/l-teles/fleetdm.svg"
brand_color: "#1A73E8"
display_name: FleetDM
name: fleetdm
description: Steampipe plugin for querying FleetDM hosts, per-host software, certificates and profiles, vulnerabilities, MDM commands and status, configuration profiles, scripts, policies, queries, labels, users, teams, global configuration, and audit activities.
og_description: Query FleetDM with SQL! Open source CLI. No DB required.
og_image: "/images/plugins/l-teles/fleetdm-social-graphic.png"
engines: ["steampipe", "sqlite", "postgres", "export"]
---

# FleetDM + Steampipe

[Steampipe](https://steampipe.io) is an open-source zero-ETL engine to instantly query cloud APIs using SQL.

[FleetDM](https://fleetdm.com) is an open-source device management platform that helps you manage and secure your devices using osquery.

For example:

```sql
select
  id,
  hostname,
  platform,
  os_version,
  status,
  team_name
from
  fleetdm_host
order by
  hostname;
```

```
+----+----------------+----------+------------+--------+----------------+
| id | hostname       | platform | os_version | status | team_name      |
+----+----------------+----------+------------+--------+----------------+
| 1  | laptop-001     | darwin   | 13.2.1     | online | Engineering    |
| 2  | server-001     | linux    | 22.04 LTS  | online | Infrastructure |
+----+----------------+----------+------------+--------+----------------+
```

## Documentation

- **[Table definitions & examples →](/plugins/l-teles/fleetdm/tables)**

## Get started

### Install

Download and install the latest FleetDM plugin:

```bash
steampipe plugin install l-teles/fleetdm
```

### Configuration

Installing the latest fleetdm plugin will create a config file (`~/.steampipe/config/fleetdm.spc`) with a single connection named `fleetdm`:

```hcl
connection "fleetdm" {
  plugin = "l-teles/fleetdm"

  # FleetDM server URL (e.g., "https://fleet.example.com")
  # The plugin will attempt to append /api/v1/ if it's not present.
  server_url = "https://fleet.example.com"

  # FleetDM API Token
  # Generate this from your FleetDM instance (User Menu -> My account -> Get API token)
  api_token = "your_api_token"

  # Optional: opt in to returning live secret material (e.g. team enrollment
  # secrets) in query results. Default: false.
  # expose_secrets = false

  # Optional: per-request timeout in seconds. Default: 30.
  # request_timeout = 30

  # Optional: cap FleetDM API requests per second. Default: no limit.
  # max_requests_per_second = 20
}
```

- `server_url` - Your FleetDM server URL. Must use `https` (`http` is only allowed for localhost). The plugin will attempt to append `/api/v1/` if it's not present.
- `api_token` - Your FleetDM API token, which can be generated from your FleetDM instance (User Menu -> My account -> Get API token)
- `expose_secrets` - Optional, default `false`. When `true`, data that can hold credentials is returned: enroll secrets (`fleetdm_enroll_secret`, `fleetdm_team.secrets`), webhook, integration and logging settings (`fleetdm_config`, `fleetdm_team`), MDM command payloads and results, and script and software install contents and outputs. When `false` those columns are NULL and `fleetdm_enroll_secret` returns an error. Leave unset unless you need it: exposed secrets will land in Steampipe query caches, exports, and dashboards.
- `max_requests_per_second` - Optional, no limit by default. Caps the number of FleetDM API requests per second for this connection, across all tables and including paging and per-host requests.
- `request_timeout` - Optional, default `30`. Per-request timeout in seconds for FleetDM API calls.

### Credentials via environment variables

Instead of setting `server_url`/`api_token` in the config file, you can use environment variables (the config file takes precedence when both are set):

```sh
export FLEETDM_URL=https://fleet.example.com
export FLEETDM_API_TOKEN=your_api_token
```

There is deliberately no environment variable for `expose_secrets` — widening secret exposure must be explicit in the config file.
