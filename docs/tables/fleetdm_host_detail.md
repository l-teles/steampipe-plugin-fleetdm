---
title: "Steampipe Table: fleetdm_host_detail - Query FleetDM Host details using SQL"
description: "Allows users to query FleetDM host details, providing insights into device status, hardware details, and system configurations across your managed fleet. Includes fields like last_restarted_at that are not part of the fleetdm_host table"
---
# Table: fleetdm_host_detail - Query FleetDM Host details using SQL

FleetDM is an open-source device management platform that helps you manage and secure your devices. The hosts table provides comprehensive information about each managed device, including system details, hardware specifications, and operational status.

## Table Usage Guide

The `fleetdm_host_detail` table provides detailed insights into your managed devices within FleetDM. As a system administrator or security analyst, you can use this table to monitor device health, track system configurations, and manage device inventory. The table helps you understand device status, hardware specifications, and operational metrics across your fleet.


## Examples

### List failing policies for a specific host
```sql
select
  h.hostname,
  p ->> 'name' as policy_name,
  p ->> 'resolution' as policy_resolution
from
  fleetdm_host_detail as h,
  jsonb_array_elements(h.policies) as p
where
  h.id = 1
  and p ->> 'response' = 'fail';
```

### Get a list of installed software for a specific host
```sql
select
  h.hostname,
  s ->> 'name' as software_name,
  s ->> 'version' as software_version,
  s ->> 'source' as software_source
from
  fleetdm_host_detail as h,
  jsonb_array_elements(h.software) as s
where
  h.id = 1;
```

### Find hosts with low battery cycle count
```sql
select
  hostname,
  platform,
  b ->> 'cycle_count' as battery_cycle_count,
  b ->> 'health' as battery_health
from
  fleetdm_host_detail,
  jsonb_array_elements(batteries) as b
where
  (b ->> 'cycle_count')::int < 50;
```

### List local users on a specific host
```sql
select
  h.hostname,
  u ->> 'username' as username,
  u ->> 'uid' as uid,
  u ->> 'shell' as shell
from
  fleetdm_host_detail h,
  jsonb_array_elements(h.users) as u
where
  h.id = 1;
```