# Table: fleetdm_host

Hosts are the individual devices (laptops, servers, etc.) managed by FleetDM. This table provides detailed information about each host.

## Columns

| Name                             | Type        | Description                                                                                         |
| -------------------------------- | ----------- | --------------------------------------------------------------------------------------------------- |
| id                               | `INT`       | The unique ID of the host.                                                                          |
| hostname                         | `TEXT`      | The hostname of the host.                                                                           |
| uuid                             | `TEXT`      | The unique UUID of the host.                                                                        |
| display_name                     | `TEXT`      | The display name of the host.                                                                       |
| display_text                     | `TEXT`      | The display text for the host (often same as hostname or display_name).                             |
| computer_name                    | `TEXT`      | The computer name of the host.                                                                      |
| osquery_host_id                  | `TEXT`      | The osquery host identifier.                                                                        |
| node_key                         | `TEXT`      | The node key for the host.                                                                          |
| status                           | `TEXT`      | The current status of the host (online, offline, mia).                                              |
| seen_time                        | `TIMESTAMP` | Timestamp when the host was last seen by Fleet.                                                     |
| created_at                       | `TIMESTAMP` | Timestamp when the host was created in Fleet.                                                       |
| updated_at                       | `TIMESTAMP` | Timestamp when the host record was last updated in Fleet.                                             |
| software_updated_at              | `TIMESTAMP` | Timestamp when the host software inventory was last updated.                                          |
| detail_updated_at                | `TIMESTAMP` | Timestamp when the host details were last updated.                                                  |
| label_updated_at                 | `TIMESTAMP` | Timestamp when the host labels were last updated.                                                   |
| policy_updated_at                | `TIMESTAMP` | Timestamp when the host policy status was last updated.                                               |
| last_enrolled_at                 | `TIMESTAMP` | Timestamp when the host last enrolled.                                                              |
| last_restarted_at                | `TIMESTAMP` | Timestamp of the last host restart event.                                                           |
| refetch_requested                | `BOOLEAN`   | Indicates if a refetch of host details has been requested.                                          |
| refetch_critical_queries_until   | `TIMESTAMP` | Timestamp until which critical queries will be refetched for this host.                             |
| platform                         | `TEXT`      | The platform of the host (e.g., 'darwin', 'windows', 'linux').                                      |
| platform_like                    | `TEXT`      | Platform-like classification (e.g., 'darwin').                                                      |
| os_version                       | `TEXT`      | The operating system version.                                                                       |
| build                            | `TEXT`      | The operating system build string.                                                                  |
| code_name                        | `TEXT`      | The OS code name.                                                                                   |
| osquery_version                  | `TEXT`      | The version of osquery running on the host.                                                         |
| orbit_version                    | `TEXT`      | The version of Orbit running on the host.                                                           |
| fleet_desktop_version            | `TEXT`      | The version of Fleet Desktop running on the host.                                                   |
| scripts_enabled                  | `BOOLEAN`   | Indicates if running scripts is enabled for this host via Fleet.                                    |
| uptime                           | `BIGINT`    | Uptime of the host in nanoseconds (as per FleetDM API).                                             |
| memory                           | `BIGINT`    | Total physical memory in bytes.                                                                     |
| cpu_type                         | `TEXT`      | CPU type.                                                                                           |
| cpu_subtype                      | `TEXT`      | CPU subtype.                                                                                        |
| cpu_brand                        | `TEXT`      | CPU brand string.                                                                                   |
| cpu_physical_cores               | `INT`       | Number of physical CPU cores.                                                                       |
| cpu_logical_cores                | `INT`       | Number of logical CPU cores.                                                                        |
| hardware_vendor                  | `TEXT`      | Hardware vendor.                                                                                    |
| hardware_model                   | `TEXT`      | Hardware model.                                                                                     |
| hardware_version                 | `TEXT`      | Hardware version.                                                                                   |
| hardware_serial                  | `TEXT`      | Hardware serial number.                                                                             |
| primary_ip                       | `IPADDR`    | The primary IP address of the host.                                                                 |
| primary_mac                      | `TEXT`      | The primary MAC address of the host.                                                                |
| public_ip                        | `IPADDR`    | The public IP address of the host.                                                                  |
| team_id                          | `INT`       | The ID of the team the host belongs to, if any.                                                     |
| team_name                        | `TEXT`      | The name of the team the host belongs to, if any.                                                   |
| distributed_interval             | `INT`       | The distributed query interval for the host.                                                        |
| config_tls_refresh               | `INT`       | The config TLS refresh interval.                                                                    |
| logger_tls_period                | `INT`       | The logger TLS period.                                                                              |
| pack_stats                       | `JSONB`     | Statistics for query packs on the host.                                                             |
| gigs_disk_space_available        | `DOUBLE`    | Gigabytes of disk space available.                                                                  |
| percent_disk_space_available     | `DOUBLE`    | Percentage of disk space available.                                                                 |
| gigs_total_disk_space            | `DOUBLE`    | Total gigabytes of disk space.                                                                      |
| issues                           | `JSONB`     | Host issues summary (failing policies, vulnerabilities).                                            |
| mdm                              | `JSONB`     | Mobile Device Management (MDM) information for the host.                                            |
| server_url                       | `TEXT`      | FleetDM server URL from connection config.                                                          |

## Example Queries

**List all online macOS hosts:**
```sql
SELECT
  id,
  hostname,
  os_version,
  team_name,
  seen_time
FROM
  fleetdm_host
WHERE
  platform = 'darwin' AND status = 'online'
ORDER BY
  hostname;
```

**Get MDM enrollment status and solution for hosts:**
```sql
SELECT
  hostname,
  platform,
  mdm ->> 'enrollment_status' AS mdm_enrollment_status,
  mdm ->> 'name' AS mdm_solution
FROM
  fleetdm_host
WHERE
  mdm IS NOT NULL;
```

**Find hosts with less than 10% disk space available:**
```sql
SELECT
  hostname,
  gigs_disk_space_available,
  percent_disk_space_available,
  gigs_total_disk_space
FROM
  fleetdm_host
WHERE
  percent_disk_space_available < 10
ORDER BY
  percent_disk_space_available;
```

**Count hosts by platform:**
```sql
SELECT
  platform,
  COUNT(*) AS host_count
FROM
  fleetdm_host
GROUP BY
  platform
ORDER BY
  host_count DESC;
