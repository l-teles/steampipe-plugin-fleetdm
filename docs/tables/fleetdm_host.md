---
title: "Steampipe Table: fleetdm_host - Query FleetDM Hosts using SQL"
description: "Allows users to query FleetDM hosts, providing insights into device status, hardware details, and system configurations across your managed fleet."
---

# Table: fleetdm_host - Query FleetDM Hosts using SQL

FleetDM is an open-source device management platform that helps you manage and secure your devices. The hosts table provides comprehensive information about each managed device, including system details, hardware specifications, and operational status.

## Table Usage Guide

The `fleetdm_host` table provides detailed insights into your managed devices within FleetDM. As a system administrator or security analyst, you can use this table to monitor device health, track system configurations, and manage device inventory. The table helps you understand device status, hardware specifications, and operational metrics across your fleet.

**Note:** This table supports the following optional key columns for server-side API filtering:

- `query` — Search hosts by hostname, hardware serial, UUID, IP address, or email.
- `team_id` — Filter hosts by team.
- `status` — Filter by host status (`new`, `online`, `offline`, `mia`, `missing`).
- `os_version_id` — Filter by OS version ID.
- `vulnerability` — Filter hosts by CVE (e.g., `CVE-2024-1234`).
- `software_version_id` — Filter hosts with a specific software version.
- `software_title_id` — Filter hosts with a specific software title.
- `policy_id` — Filter hosts by policy evaluation results (use with `policy_response`).
- `policy_response` — Filter by `passing` or `failing` for the specified `policy_id`.
- `mdm_enrollment_status` — Filter by MDM enrollment status.
- `low_disk_space` — Filter hosts with less than this number of GB free (Fleet Premium, 1–100).

Using these key columns in your `WHERE` clause pushes the filtering to the FleetDM API, reducing data transfer and improving query performance.

## Examples

### List all online macOS hosts

Identify all currently online macOS devices to monitor their status and configurations.

```sql+postgres
select
  id,
  hostname,
  os_version,
  team_name,
  seen_time
from
  fleetdm_host
where
  platform = 'darwin' and status = 'online'
order by
  hostname;
```

```sql+sqlite
select
  id,
  hostname,
  os_version,
  team_name,
  seen_time
from
  fleetdm_host
where
  platform = 'darwin' and status = 'online'
order by
  hostname;
```

### Get MDM enrollment status and solution for hosts

Monitor Mobile Device Management enrollment status across your fleet to ensure proper device management.

```sql+postgres
select
  hostname,
  platform,
  mdm ->> 'enrollment_status' as mdm_enrollment_status,
  mdm ->> 'name' as mdm_solution
from
  fleetdm_host
where
  mdm is not null;
```

```sql+sqlite
select
  hostname,
  platform,
  json_extract(mdm, '$.enrollment_status') as mdm_enrollment_status,
  json_extract(mdm, '$.name') as mdm_solution
from
  fleetdm_host
where
  mdm is not null;
```

### Find hosts with less than 10% disk space available

Identify devices with critical disk space issues that require immediate attention.

```sql+postgres
select
  hostname,
  gigs_disk_space_available,
  percent_disk_space_available,
  gigs_total_disk_space
from
  fleetdm_host
where
  percent_disk_space_available < 10
order by
  percent_disk_space_available;
```

```sql+sqlite
select
  hostname,
  gigs_disk_space_available,
  percent_disk_space_available,
  gigs_total_disk_space
from
  fleetdm_host
where
  percent_disk_space_available < 10
order by
  percent_disk_space_available;
```

### Search hosts by hostname or serial number

Use the `query` key column to search hosts server-side by hostname, hardware serial, UUID, IP address, or email.

```sql+postgres
select
  id,
  hostname,
  platform,
  status
from
  fleetdm_host
where
  query = 'macbook-pro'
order by
  hostname;
```

```sql+sqlite
select
  id,
  hostname,
  platform,
  status
from
  fleetdm_host
where
  query = 'macbook-pro'
order by
  hostname;
```

### Find hosts affected by a specific CVE

Use the `vulnerability` key column to find all hosts affected by a given CVE.

```sql+postgres
select
  id,
  hostname,
  platform,
  os_version
from
  fleetdm_host
where
  vulnerability = 'CVE-2024-3596';
```

```sql+sqlite
select
  id,
  hostname,
  platform,
  os_version
from
  fleetdm_host
where
  vulnerability = 'CVE-2024-3596';
```

### List hosts failing a specific policy

Use `policy_id` with `policy_response` to find hosts that are failing a particular policy.

```sql+postgres
select
  id,
  hostname,
  team_name
from
  fleetdm_host
where
  policy_id = 42
  and policy_response = 'failing';
```

```sql+sqlite
select
  id,
  hostname,
  team_name
from
  fleetdm_host
where
  policy_id = 42
  and policy_response = 'failing';
```

### Count hosts by platform

Analyze the distribution of operating systems across your fleet to understand platform diversity.

```sql+postgres
select
  platform,
  count(*) as host_count
from
  fleetdm_host
group by
  platform
order by
  host_count desc;
```

```sql+sqlite
select
  platform,
  count(*) as host_count
from
  fleetdm_host
group by
  platform
order by
  host_count desc;
```
