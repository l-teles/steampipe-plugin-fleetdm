---
title: "Steampipe Table: fleetdm_software_title - Query FleetDM Software Titles using SQL"
description: "Allows users to query FleetDM software titles, providing insights into software groupings, version counts, and vulnerability information across managed hosts."
---

# Table: fleetdm_software_title - Query FleetDM Software Titles using SQL

FleetDM is an open-source device management platform that helps you manage and secure your devices. The software titles table provides information about software titles across your managed hosts. A software title groups multiple versions of the same software, giving you an aggregated view of software distribution and deployment. Uses the `/software/titles` API endpoint.

## Table Usage Guide

The `fleetdm_software_title` table provides aggregated insights into your software inventory within FleetDM. As a system administrator or security analyst, you can use this table to understand which software titles are installed across your fleet, how many versions exist for each, and which titles have associated vulnerabilities. This is particularly useful for software lifecycle management and identifying software available for install.

> **Note:** The FleetDM API requires `vulnerable=true` when using `min_cvss_score`, `max_cvss_score`, or `exploit` filters. The plugin automatically sets `vulnerable=true` when any of these are specified, so you don't need to include `vulnerable_only = true` explicitly in those queries (though it's still recommended for clarity).

## Columns

| Name                          | Type      | Description                                                                                                                                        |
| ----------------------------- | --------- | -------------------------------------------------------------------------------------------------------------------------------------------------- |
| id                            | `INT`     | Unique ID of the software title.                                                                                                                   |
| name                          | `TEXT`    | Name of the software title.                                                                                                                        |
| display_name                  | `TEXT`    | Display name of the software title.                                                                                                                |
| icon_url                      | `TEXT`    | URL of the software icon.                                                                                                                          |
| source                        | `TEXT`    | Source of the software information (e.g., 'apps', 'deb_packages', 'chrome_extensions').                                                            |
| extension_for                 | `TEXT`    | If a browser extension, specifies which software it extends.                                                                                       |
| browser                       | `TEXT`    | Browser name for browser extensions.                                                                                                               |
| hosts_count                   | `INT`     | Number of hosts where this software title is installed.                                                                                            |
| versions_count                | `INT`     | Number of distinct versions of this software title.                                                                                                |
| bundle_identifier             | `TEXT`    | Bundle identifier, typically for macOS and iOS software.                                                                                           |
| versions                      | `JSONB`   | List of versions for this software title, including version IDs and associated vulnerabilities.                                                    |
| software_package              | `JSONB`   | Software package details if the software was added for install.                                                                                    |
| app_store_app                 | `JSONB`   | App Store app details if the software is from an app store.                                                                                        |
| vulnerable_only               | `BOOLEAN` | (Key Column) Filter for software titles with known vulnerabilities. Use in `WHERE` clause.                                                         |
| team_id                       | `INT`     | (Key Column) Filter by team ID (Fleet Premium). Use 0 for hosts assigned to 'No team'. Use in `WHERE` clause.                                      |
| available_for_install         | `BOOLEAN` | (Key Column) Filter for software available for install (added by the user). Use in `WHERE` clause.                                                 |
| query                         | `TEXT`    | (Key Column) Search query keywords. Searchable fields include title and CVE. Use in `WHERE` clause.                                                |
| self_service                  | `BOOLEAN` | (Key Column) Filter for self-service software only. Use in `WHERE` clause.                                                                         |
| packages_only                 | `BOOLEAN` | (Key Column) Filter for install packages only, excluding app store apps (Fleet Premium). Use in `WHERE` clause.                                    |
| min_cvss_score                | `INT`     | (Key Column) Filter for software with vulnerabilities having a CVSS v3.x base score higher than this value (Fleet Premium). Use in `WHERE` clause. |
| max_cvss_score                | `INT`     | (Key Column) Filter for software with vulnerabilities having a CVSS v3.x base score lower than this value (Fleet Premium). Use in `WHERE` clause.  |
| exploit                       | `BOOLEAN` | (Key Column) Filter for software with CISA-known actively exploited vulnerabilities (Fleet Premium). Use in `WHERE` clause.                        |
| platform                      | `TEXT`    | (Key Column) Filter installable titles by platform. Options: 'macos', 'darwin', 'windows', 'linux', 'chrome', 'ios', 'ipados'. Requires team_id.   |
| exclude_fleet_maintained_apps | `BOOLEAN` | (Key Column) Exclude Fleet-maintained apps from the results. Use in `WHERE` clause.                                                                |
| server_url                    | `TEXT`    | FleetDM server URL from connection config.                                                                                                         |

## Examples

### List top 10 most common software titles

Identify the most widely deployed software titles across your fleet to help with standardization and license management.

```sql+postgres
select
  id,
  name,
  source,
  hosts_count,
  versions_count
from
  fleetdm_software_title
order by
  hosts_count desc
limit 10;
```

```sql+sqlite
select
  id,
  name,
  source,
  hosts_count,
  versions_count
from
  fleetdm_software_title
order by
  hosts_count desc
limit 10;
```

### Find software titles with multiple versions installed

Identify software titles that have many different versions deployed, which may indicate inconsistent update practices.

```sql+postgres
select
  name,
  display_name,
  source,
  versions_count,
  hosts_count
from
  fleetdm_software_title
where
  versions_count > 1
order by
  versions_count desc;
```

```sql+sqlite
select
  name,
  display_name,
  source,
  versions_count,
  hosts_count
from
  fleetdm_software_title
where
  versions_count > 1
order by
  versions_count desc;
```

### List vulnerable software titles

Identify software titles with known vulnerabilities to prioritize remediation efforts.

```sql+postgres
select
  name,
  display_name,
  hosts_count,
  versions_count
from
  fleetdm_software_title
where
  vulnerable_only = true
order by
  hosts_count desc;
```

```sql+sqlite
select
  name,
  display_name,
  hosts_count,
  versions_count
from
  fleetdm_software_title
where
  vulnerable_only = 1
order by
  hosts_count desc;
```

### List software available for install on a specific team

Find software titles that are available for installation on a given team.

```sql+postgres
select
  name,
  display_name,
  source,
  hosts_count
from
  fleetdm_software_title
where
  available_for_install = true
  and team_id = 1
order by
  name;
```

```sql+sqlite
select
  name,
  display_name,
  source,
  hosts_count
from
  fleetdm_software_title
where
  available_for_install = 1
  and team_id = 1
order by
  name;
```

### Extract version details and vulnerabilities for a software title

Get detailed version and vulnerability information for a specific software title.

```sql+postgres
select
  st.name,
  v ->> 'id' as version_id,
  v ->> 'version' as version,
  v ->> 'hosts_count' as hosts_count,
  v -> 'vulnerabilities' as vulnerabilities
from
  fleetdm_software_title as st,
  jsonb_array_elements(st.versions) as v
where
  st.name = 'Google Chrome.app';
```

```sql+sqlite
select
  st.name,
  json_extract(v.value, '$.id') as version_id,
  json_extract(v.value, '$.version') as version,
  json_extract(v.value, '$.hosts_count') as hosts_count,
  json_extract(v.value, '$.vulnerabilities') as vulnerabilities
from
  fleetdm_software_title as st,
  json_each(st.versions) as v
where
  st.name = 'Google Chrome.app';
```

### Search software titles by keyword

Use the `query` key column to search across title names and CVEs.

```sql+postgres
select
  name,
  display_name,
  hosts_count,
  versions_count
from
  fleetdm_software_title
where
  query = 'Chrome';
```

```sql+sqlite
select
  name,
  display_name,
  hosts_count,
  versions_count
from
  fleetdm_software_title
where
  query = 'Chrome';
```

### Find self-service software titles

List software titles that are available as self-service for end users.

```sql+postgres
select
  name,
  display_name,
  source,
  hosts_count
from
  fleetdm_software_title
where
  self_service = true
order by
  name;
```

```sql+sqlite
select
  name,
  display_name,
  source,
  hosts_count
from
  fleetdm_software_title
where
  self_service = 1
order by
  name;
```

### Find software titles with high-severity vulnerabilities (CVSS ≥ 9)

Filter for software titles with critical vulnerabilities based on CVSS v3.x base score (Fleet Premium).

```sql+postgres
select
  name,
  display_name,
  hosts_count,
  versions_count
from
  fleetdm_software_title
where
  min_cvss_score = 9;
```

```sql+sqlite
select
  name,
  display_name,
  hosts_count,
  versions_count
from
  fleetdm_software_title
where
  min_cvss_score = 9;
```

### List macOS software titles available for install on a team

Filter installable software titles by platform (requires `team_id`).

```sql+postgres
select
  name,
  display_name,
  source,
  hosts_count
from
  fleetdm_software_title
where
  available_for_install = true
  and team_id = 1
  and platform = 'darwin'
order by
  name;
```

```sql+sqlite
select
  name,
  display_name,
  source,
  hosts_count
from
  fleetdm_software_title
where
  available_for_install = 1
  and team_id = 1
  and platform = 'darwin'
order by
  name;
```
