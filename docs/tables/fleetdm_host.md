---
title: "Steampipe Table: fleetdm_host - Query FleetDM Hosts using SQL"
description: "Allows users to query FleetDM hosts, providing insights into device status, hardware details, and system configurations across your managed fleet."
---

# Table: fleetdm_host - Query FleetDM Hosts using SQL

FleetDM is an open-source device management platform that helps you manage and secure your devices. The hosts table provides comprehensive information about each managed device, including system details, hardware specifications, and operational status.

## Table Usage Guide

The `fleetdm_host` table provides detailed insights into your managed devices within FleetDM. As a system administrator or security analyst, you can use this table to monitor device health, track system configurations, and manage device inventory. The table helps you understand device status, hardware specifications, and operational metrics across your fleet.

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
