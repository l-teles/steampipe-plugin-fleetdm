---
title: "Steampipe Table: fleetdm_script - Query FleetDM Scripts using SQL"
description: "Allows users to query the scripts saved in FleetDM across all teams."
---

# Table: fleetdm_script - Query FleetDM Scripts using SQL

FleetDM can run saved scripts on hosts, on demand or as policy automations. This table lists saved scripts for hosts with no team and for every team. Uses the `/scripts` API endpoint; script contents are not returned.

## Table Usage Guide

The `fleetdm_script` table gives an inventory of the scripts available to run on hosts. Without a `team_id` filter it lists scripts for hosts with no team plus every team (one request per team); with `team_id` it lists only that team's scripts.

## Examples

### List all scripts by team

See which scripts each team has.

```sql+postgres
select
  s.name,
  s.team_id,
  t.name as team_name,
  s.updated_at
from
  fleetdm_script s
  left join fleetdm_team t on t.id = s.team_id
order by
  t.name,
  s.name;
```

```sql+sqlite
select
  s.name,
  s.team_id,
  t.name as team_name,
  s.updated_at
from
  fleetdm_script s
  left join fleetdm_team t on t.id = s.team_id
order by
  t.name,
  s.name;
```

### List a team's scripts



```sql+postgres
select
  id,
  name,
  updated_at
from
  fleetdm_script
where
  team_id = 1;
```

```sql+sqlite
select
  id,
  name,
  updated_at
from
  fleetdm_script
where
  team_id = 1;
```
