# Table: fleetdm_software

This table lists all software items inventoried by FleetDM across all managed hosts, including version, source, and vulnerability information.

## Columns

| Name                             | Type        | Description                                                                                                |
| -------------------------------- | ----------- | ---------------------------------------------------------------------------------------------------------- |
| id                               | `INT`       | Unique ID of the software item.                                                                            |
| name                             | `TEXT`      | Name of the software.                                                                                      |
| version                          | `TEXT`      | Version of the software.                                                                                   |
| source                           | `TEXT`      | Source of the software information (e.g., 'apps', 'deb_packages', 'chrome_extensions', 'rpm_packages').    |
| host_count                       | `INT`       | Number of hosts where this software is installed.                                                          |
| generated_cpe                    | `TEXT`      | Generated Common Platform Enumeration (CPE) string for the software.                                       |
| bundle_identifier                | `TEXT`      | Bundle identifier, typically for macOS and iOS software.                                                   |
| release                          | `TEXT`      | Release information, e.g., for RPM packages.                                                               |
| vendor                           | `TEXT`      | Vendor information, e.g., for RPM packages.                                                                |
| arch                             | `TEXT`      | Architecture information, e.g., for RPM packages.                                                          |
| extension_id                     | `TEXT`      | Extension ID for browser extensions.                                                                       |
| browser                          | `TEXT`      | Browser name for browser extensions.                                                                       |
| path                             | `TEXT`      | Install path for certain software types like Programs.                                                     |
| installed_path                   | `TEXT`      | Installed path, e.g., for Homebrew packages.                                                               |
| last_opened_at                   | `TIMESTAMP` | Timestamp when the software was last opened (may be aggregated or host-specific).                          |
| counts_updated_at                | `TIMESTAMP` | Timestamp when the host_count for this software item was last updated.                                     |
| vulnerabilities                  | `JSONB`     | Vulnerabilities associated with this software. Contains an array of vulnerability objects.                 |
| vulnerable_only                  | `BOOLEAN`   | (Key Column) Filter for software with known vulnerabilities. Use in `WHERE` clause.                        |
| os_id                            | `INT`       | (Key Column) Filter by OS ID. Use in `WHERE` clause.                                                       |
| os_name                          | `TEXT`      | (Key Column) Filter by OS name (e.g. 'Ubuntu', 'Windows Server 2019 Datacenter'). Use in `WHERE` clause. |
| os_version                       | `TEXT`      | (Key Column) Filter by OS version (e.g. '20.04.4 LTS', '10.0.17763'). Use in `WHERE` clause.             |
| team_id                          | `INT`       | (Key Column) Filter by team ID. Use in `WHERE` clause.                                                     |
| server_url                       | `TEXT`      | FleetDM server URL from connection config.                                                                 |

## Example Queries

**List top 10 most common software versions:**
```sql
SELECT
  name,
  version,
  source,
  host_count
FROM
  fleetdm_software
ORDER BY
  host_count DESC
LIMIT 10;
```

**Find all versions of Google Chrome installed:**
```sql
SELECT
  name,
  version,
  host_count
FROM
  fleetdm_software
WHERE
  name = 'Google Chrome.app' -- Or appropriate name based on your FleetDM data
ORDER BY
  version;
```

**List software with known vulnerabilities:**
```sql
SELECT
  name,
  version,
  host_count,
  jsonb_array_length(vulnerabilities) AS vulnerability_count
FROM
  fleetdm_software
WHERE
  vulnerabilities IS NOT NULL
ORDER BY
  vulnerability_count DESC,
  host_count DESC;
```

**Extract CVEs for a specific software item:**
```sql
SELECT
  s.name,
  s.version,
  v ->> 'cve' AS cve,
  v ->> 'details_link' AS details_link
FROM
  fleetdm_software AS s,
  jsonb_array_elements(s.vulnerabilities) AS v
WHERE
  s.name = 'Firefox.app' AND s.version = '100.0.1'; -- Adjust to a software item in your inventory
```