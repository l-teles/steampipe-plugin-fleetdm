---
title: "Steampipe Table: fleetdm_host_summary - Query FleetDM Host Summary using SQL"
description: "Allows users to query FleetDM's host count summary."
---

# Table: fleetdm_host_summary - Query FleetDM Host Summary using SQL

The host summary counts hosts by status and platform, as on Fleet's dashboard. Uses the `/host_summary` API endpoint.

## Table Usage Guide

The `fleetdm_host_summary` table returns one row. Filter with `team_id` (0 for hosts with no team), `platform` or `low_disk_space`; the filters are sent to the API.

## Examples

### Get online and offline host counts

```sql+postgres
select
  totals_hosts_count,
  online_count,
  offline_count,
  mia_count
from
  fleetdm_host_summary;
```

```sql+sqlite
select
  totals_hosts_count,
  online_count,
  offline_count,
  mia_count
from
  fleetdm_host_summary;
```

### Count hosts with less than 10 GB free disk space

```sql+postgres
select
  low_disk_space_count
from
  fleetdm_host_summary
where
  low_disk_space = 10;
```

```sql+sqlite
select
  low_disk_space_count
from
  fleetdm_host_summary
where
  low_disk_space = 10;
```
