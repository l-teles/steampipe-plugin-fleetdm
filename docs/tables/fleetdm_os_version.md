---
title: "Steampipe Table: fleetdm_os_version - Query FleetDM OS Versions using SQL"
description: "Allows users to query FleetDM operating system versions, providing insights into OS distribution, host counts, and vulnerability information across managed hosts."
---

# Table: fleetdm_os_version - Query FleetDM OS Versions using SQL

FleetDM is an open-source device management platform that helps you manage and secure your devices. The OS versions table provides comprehensive information about operating system versions deployed across your managed hosts, including host counts, platform details, and associated vulnerabilities. Uses the `/os_versions` API endpoint.

## Table Usage Guide

The `fleetdm_os_version` table provides detailed insights into the operating system landscape across your fleet. As a system administrator or security analyst, you can use this table to track OS version distribution, identify outdated operating systems, and discover OS-level vulnerabilities. The table helps you ensure compliance with OS version requirements and prioritize OS upgrade efforts.

## Examples

### List all OS versions ordered by host count

Get an overview of OS version distribution across your fleet to understand your most common operating systems.

```sql+postgres
select
  os_version_id,
  name,
  version,
  platform,
  hosts_count,
  vulnerabilities_count
from
  fleetdm_os_version
order by
  hosts_count desc;
```

```sql+sqlite
select
  os_version_id,
  name,
  version,
  platform,
  hosts_count,
  vulnerabilities_count
from
  fleetdm_os_version
order by
  hosts_count desc;
```

### Find OS versions with known vulnerabilities

Identify operating system versions that have known vulnerabilities to prioritize patching and upgrades.

```sql+postgres
select
  name,
  version,
  platform,
  hosts_count,
  vulnerabilities_count
from
  fleetdm_os_version
where
  vulnerabilities_count > 0
order by
  vulnerabilities_count desc,
  hosts_count desc;
```

```sql+sqlite
select
  name,
  version,
  platform,
  hosts_count,
  vulnerabilities_count
from
  fleetdm_os_version
where
  vulnerabilities_count > 0
order by
  vulnerabilities_count desc,
  hosts_count desc;
```

### List OS versions by platform

Get a breakdown of OS versions for a specific platform to track version standardization.

```sql+postgres
select
  name,
  version,
  hosts_count,
  vulnerabilities_count
from
  fleetdm_os_version
where
  platform = 'darwin'
order by
  hosts_count desc;
```

```sql+sqlite
select
  name,
  version,
  hosts_count,
  vulnerabilities_count
from
  fleetdm_os_version
where
  platform = 'darwin'
order by
  hosts_count desc;
```

### Extract CVE details for a specific OS version

Get detailed vulnerability information for a specific operating system version to assess security risks.

```sql+postgres
select
  o.name,
  o.version,
  v ->> 'cve' as cve,
  v ->> 'details_link' as details_link,
  v ->> 'cvss_score' as cvss_score,
  v ->> 'resolved_in_version' as resolved_in_version
from
  fleetdm_os_version as o,
  jsonb_array_elements(o.vulnerabilities) as v
where
  o.name_only = 'macOS'
order by
  (v ->> 'cvss_score')::float desc nulls last;
```

```sql+sqlite
select
  o.name,
  o.version,
  json_extract(v.value, '$.cve') as cve,
  json_extract(v.value, '$.details_link') as details_link,
  json_extract(v.value, '$.cvss_score') as cvss_score,
  json_extract(v.value, '$.resolved_in_version') as resolved_in_version
from
  fleetdm_os_version as o,
  json_each(o.vulnerabilities) as v
where
  o.name_only = 'macOS'
order by
  cast(json_extract(v.value, '$.cvss_score') as real) desc;
```

### List OS versions for a specific team

Filter OS versions by team to understand the OS landscape within a specific group.

```sql+postgres
select
  name,
  version,
  platform,
  hosts_count
from
  fleetdm_os_version
where
  team_id = 1
order by
  hosts_count desc;
```

```sql+sqlite
select
  name,
  version,
  platform,
  hosts_count
from
  fleetdm_os_version
where
  team_id = 1
order by
  hosts_count desc;
```
