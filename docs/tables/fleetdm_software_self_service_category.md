---
title: "Steampipe Table: fleetdm_software_self_service_category - Query FleetDM Self-Service Software Categories using SQL"
description: "Allows users to query the categories that group self-service software in Fleet Desktop."
---

# Table: fleetdm_software_self_service_category - Query FleetDM Self-Service Software Categories using SQL

Self-service categories group the software end users can install from Fleet Desktop. Uses the `/software/self_service_categories` API endpoint.

## Table Usage Guide

Without a `team_id` filter the `fleetdm_software_self_service_category` table lists categories for hosts with no team (`team_id = 0`) plus every team, one request per team.

## Examples

### Count categories per team

```sql+postgres
select
  team_id,
  count(*) as categories
from
  fleetdm_software_self_service_category
group by
  team_id
order by
  team_id;
```

```sql+sqlite
select
  team_id,
  count(*) as categories
from
  fleetdm_software_self_service_category
group by
  team_id
order by
  team_id;
```
