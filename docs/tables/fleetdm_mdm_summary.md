---
title: "Steampipe Table: fleetdm_mdm_summary - Query FleetDM MDM Summary using SQL"
description: "Allows users to query FleetDM's MDM enrollment summary."
---

# Table: fleetdm_mdm_summary - Query FleetDM MDM Summary using SQL

The MDM summary counts hosts by MDM enrollment status. Uses the `/hosts/summary/mdm` API endpoint.

## Table Usage Guide

The `fleetdm_mdm_summary` table returns one row. Filter with `team_id` (0 for hosts with no team) or `platform`; the filters are sent to the API.

## Examples

### Get MDM enrollment counts for macOS

```sql+postgres
select
  hosts_count,
  enrolled_manual_hosts_count,
  enrolled_automated_hosts_count,
  unenrolled_hosts_count
from
  fleetdm_mdm_summary
where
  platform = 'darwin';
```

```sql+sqlite
select
  hosts_count,
  enrolled_manual_hosts_count,
  enrolled_automated_hosts_count,
  unenrolled_hosts_count
from
  fleetdm_mdm_summary
where
  platform = 'darwin';
```
