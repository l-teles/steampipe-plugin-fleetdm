---
title: "Steampipe Table: fleetdm_software_version - Query FleetDM Software Versions using SQL"
description: "Allows users to query FleetDM software version inventory, providing insights into installed software, versions, and vulnerability information across managed hosts."
---

# Table: fleetdm_software_version - Query FleetDM Software Versions using SQL

FleetDM is an open-source device management platform that helps you manage and secure your devices. The software versions table provides comprehensive information about all software version items installed across your managed hosts, including version details, installation sources, and associated vulnerabilities. Uses the `/software/versions` API endpoint.

## Table Usage Guide

The `fleetdm_software_version` table provides detailed insights into your software version inventory within FleetDM. As a system administrator or security analyst, you can use this table to track software versions, identify vulnerable software, and maintain an up-to-date inventory of installed applications across your fleet. The table helps you understand software distribution, version patterns, and security risks associated with installed software.

> **Note:** The FleetDM API requires `vulnerable=true` when using `min_cvss_score`, `max_cvss_score`, or `exploit` filters. The plugin automatically sets `vulnerable=true` when any of these are specified, so you don't need to include `vulnerable_only = true` explicitly in those queries (though it's still recommended for clarity).

## Columns

| Name              | Type        | Description                                                                                                                                                      |
| ----------------- | ----------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| id                | `INT`       | Unique ID of the software item.                                                                                                                                  |
| name              | `TEXT`      | Name of the software.                                                                                                                                            |
| version           | `TEXT`      | Version of the software.                                                                                                                                         |
| source            | `TEXT`      | Source of the software information (e.g., 'apps', 'deb_packages', 'chrome_extensions', 'rpm_packages').                                                          |
| host_count        | `INT`       | Number of hosts where this software is installed.                                                                                                                |
| generated_cpe     | `TEXT`      | Generated Common Platform Enumeration (CPE) string for the software.                                                                                             |
| bundle_identifier | `TEXT`      | Bundle identifier, typically for macOS and iOS software.                                                                                                         |
| release           | `TEXT`      | Release information, e.g., for RPM packages.                                                                                                                     |
| vendor            | `TEXT`      | Vendor information, e.g., for RPM packages.                                                                                                                      |
| arch              | `TEXT`      | Architecture information, e.g., for RPM packages.                                                                                                                |
| extension_id      | `TEXT`      | Extension ID for browser extensions.                                                                                                                             |
| browser           | `TEXT`      | Browser name for browser extensions.                                                                                                                             |
| path              | `TEXT`      | Install path for certain software types like Programs.                                                                                                           |
| installed_path    | `TEXT`      | Installed path, e.g., for Homebrew packages.                                                                                                                     |
| last_opened_at    | `TIMESTAMP` | Timestamp when the software was last opened (may be aggregated or host-specific).                                                                                |
| counts_updated_at | `TIMESTAMP` | Timestamp when the host_count for this software item was last updated.                                                                                           |
| vulnerabilities   | `JSONB`     | Vulnerabilities associated with this software. Contains an array of vulnerability objects.                                                                       |
| vulnerable_only   | `BOOLEAN`   | (Key Column) Filter for software with known vulnerabilities. Use in `WHERE` clause.                                                                              |
| team_id           | `INT`       | (Key Column) Filter by team ID (Fleet Premium). Use 0 for hosts assigned to 'No team'. Use in `WHERE` clause.                                                    |
| query             | `TEXT`      | (Key Column) Search query keywords. Searchable fields include name, version, and CVE. Use in `WHERE` clause.                                                     |
| min_cvss_score    | `INT`       | (Key Column) Filter for software with vulnerabilities having a CVSS v3.x base score higher than this value (Fleet Premium). Use in `WHERE` clause.               |
| max_cvss_score    | `INT`       | (Key Column) Filter for software with vulnerabilities having a CVSS v3.x base score lower than this value (Fleet Premium). Use in `WHERE` clause.                |
| exploit           | `BOOLEAN`   | (Key Column) Filter for software with vulnerabilities that have been actively exploited in the wild — CISA known exploit (Fleet Premium). Use in `WHERE` clause. |
| server_url        | `TEXT`      | FleetDM server URL from connection config.                                                                                                                       |

## Examples

### List top 10 most common software versions

Identify the most widely deployed software versions across your fleet to help with standardization and update planning.

```sql+postgres
select
  name,
  version,
  source,
  host_count
from
  fleetdm_software_version
order by
  host_count desc
limit 10;
```

```sql+sqlite
select
  name,
  version,
  source,
  host_count
from
  fleetdm_software_version
order by
  host_count desc
limit 10;
```

### Find all versions of Google Chrome installed

Track different versions of Google Chrome across your fleet to identify outdated installations.

```sql+postgres
select
  name,
  version,
  host_count
from
  fleetdm_software_version
where
  name = 'Google Chrome.app'
order by
  version;
```

```sql+sqlite
select
  name,
  version,
  host_count
from
  fleetdm_software_version
where
  name = 'Google Chrome.app'
order by
  version;
```

### List software with known vulnerabilities

Identify software items with known security vulnerabilities to prioritize updates and remediation efforts.

```sql+postgres
select
  name,
  version,
  host_count,
  jsonb_array_length(vulnerabilities) as vulnerability_count
from
  fleetdm_software_version
where
  vulnerabilities is not null
order by
  vulnerability_count desc,
  host_count desc;
```

```sql+sqlite
select
  name,
  version,
  host_count,
  json_array_length(vulnerabilities) as vulnerability_count
from
  fleetdm_software_version
where
  vulnerabilities is not null
order by
  vulnerability_count desc,
  host_count desc;
```

### Extract CVEs for a specific software item

Get detailed vulnerability information for a specific software version to assess security risks.

```sql+postgres
select
  s.name,
  s.version,
  v ->> 'cve' as cve,
  v ->> 'details_link' as details_link
from
  fleetdm_software_version as s,
  jsonb_array_elements(s.vulnerabilities) as v
where
  s.name = 'Firefox.app' and s.version = '100.0.1';
```

```sql+sqlite
select
  s.name,
  s.version,
  json_extract(v.value, '$.cve') as cve,
  json_extract(v.value, '$.details_link') as details_link
from
  fleetdm_software_version as s,
  json_each(s.vulnerabilities) as v
where
  s.name = 'Firefox.app' and s.version = '100.0.1';
```

### Search software by name, version, or CVE keyword

Use the `query` key column to search across software name, version, and CVE fields.

```sql+postgres
select
  name,
  version,
  host_count,
  generated_cpe
from
  fleetdm_software_version
where
  query = 'Chrome';
```

```sql+sqlite
select
  name,
  version,
  host_count,
  generated_cpe
from
  fleetdm_software_version
where
  query = 'Chrome';
```

### Find software with high-severity vulnerabilities (CVSS ≥ 9)

Filter for software with critical vulnerabilities based on CVSS v3.x base score (Fleet Premium).

```sql+postgres
select
  name,
  version,
  host_count,
  vulnerabilities
from
  fleetdm_software_version
where
  vulnerable_only = true
  and min_cvss_score = 9;
```

```sql+sqlite
select
  name,
  version,
  host_count,
  vulnerabilities
from
  fleetdm_software_version
where
  vulnerable_only = true
  and min_cvss_score = 9;
```

### List software with actively exploited vulnerabilities

Identify software with CISA-known actively exploited vulnerabilities (Fleet Premium).

```sql+postgres
select
  name,
  version,
  host_count,
  vulnerabilities
from
  fleetdm_software_version
where
  exploit = true;
```

```sql+sqlite
select
  name,
  version,
  host_count,
  vulnerabilities
from
  fleetdm_software_version
where
  exploit = true;
```
