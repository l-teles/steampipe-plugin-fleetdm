---
title: "Steampipe Table: fleetdm_pack - Query FleetDM Query Packs using SQL"
description: "Allows users to query FleetDM query packs, providing insights into scheduled queries, target configurations, and pack distributions within your FleetDM instance."
---

# Table: fleetdm_pack - Query FleetDM Query Packs using SQL

FleetDM is an open-source device management platform that helps you manage and secure your devices. Query packs in FleetDM are collections of queries that can be scheduled to run against targeted hosts or labels, enabling automated data collection and monitoring.

## Table Usage Guide

The `fleetdm_pack` table provides comprehensive insights into query pack configurations within your FleetDM instance. As a system administrator, you can use this table to manage scheduled queries, monitor pack distributions, and maintain target configurations. The table helps you understand how queries are organized and scheduled across your fleet.

## Examples

### List all query packs
Get an overview of all query packs in your FleetDM instance, including their type and target information.

```sql+postgres
select
  id,
  name,
  type,
  platform,
  disabled,
  target_count,
  total_scheduled_queries_count
from
  fleetdm_pack
order by
  name;
```

```sql+sqlite
select
  id,
  name,
  type,
  platform,
  disabled,
  target_count,
  total_scheduled_queries_count
from
  fleetdm_pack
order by
  name;
```

### Find disabled packs
Identify query packs that are currently disabled and may need review or reconfiguration.

```sql+postgres
select
  id,
  name,
  type,
  platform
from
  fleetdm_pack
where
  disabled = true;
```

```sql+sqlite
select
  id,
  name,
  type,
  platform
from
  fleetdm_pack
where
  disabled = true;
```

### List packs targeted at a specific platform
Find query packs that are configured to run on a specific operating system platform.

```sql+postgres
select
  id,
  name,
  description
from
  fleetdm_pack
where
  platform = 'darwin' or platform = '' or platform is null
order by
  name;
```

```sql+sqlite
select
  id,
  name,
  description
from
  fleetdm_pack
where
  platform = 'darwin' or platform = '' or platform is null
order by
  name;
```

### Get details for a specific pack
Examine the scheduled queries and configuration details for a particular query pack.

```sql+postgres
select
  name,
  scheduled_queries
from
  fleetdm_pack
where
  id = 1;
```

```sql+sqlite
select
  name,
  scheduled_queries
from
  fleetdm_pack
where
  id = 1;
```
