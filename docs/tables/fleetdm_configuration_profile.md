---
title: "Steampipe Table: fleetdm_configuration_profile - Query FleetDM Configuration Profiles using SQL"
description: "Allows users to query the MDM configuration profiles (OS settings) in FleetDM across all teams."
---

# Table: fleetdm_configuration_profile - Query FleetDM Configuration Profiles using SQL

FleetDM delivers configuration profiles to Apple, Windows and Android hosts through MDM. This table lists the profiles for hosts with no team and for every team. Uses the `/configuration_profiles` API endpoint; profile contents are not returned.

## Table Usage Guide

The `fleetdm_configuration_profile` table helps you review which OS settings each team enforces. Without a `team_id` filter it lists profiles for hosts with no team plus every team (one request per team); with `team_id` it lists only that team's profiles. Profiles for hosts with no team have `team_id = 0`. When MDM is not turned on in Fleet, the table returns no rows.

## Examples

### List profiles by team and platform



```sql+postgres
select
  p.name,
  p.platform,
  p.identifier,
  t.name as team_name
from
  fleetdm_configuration_profile p
  left join fleetdm_team t on t.id = p.team_id
order by
  t.name,
  p.platform,
  p.name;
```

```sql+sqlite
select
  p.name,
  p.platform,
  p.identifier,
  t.name as team_name
from
  fleetdm_configuration_profile p
  left join fleetdm_team t on t.id = p.team_id
order by
  t.name,
  p.platform,
  p.name;
```

### Find profiles scoped with labels

Profiles that only target, or exclude, some hosts.

```sql+postgres
select
  name,
  platform,
  labels_include_all,
  labels_include_any,
  labels_exclude_any
from
  fleetdm_configuration_profile
where
  labels_include_all is not null
  or labels_include_any is not null
  or labels_exclude_any is not null;
```

```sql+sqlite
select
  name,
  platform,
  labels_include_all,
  labels_include_any,
  labels_exclude_any
from
  fleetdm_configuration_profile
where
  labels_include_all is not null
  or labels_include_any is not null
  or labels_exclude_any is not null;
```
