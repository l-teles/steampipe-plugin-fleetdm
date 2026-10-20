---
title: "Steampipe Table: fleetdm_setup_experience_software - Query FleetDM Setup Experience Software using SQL"
description: "Allows users to query the software available during each team's setup experience."
---

# Table: fleetdm_setup_experience_software - Query FleetDM Setup Experience Software using SQL

Setup experience software is installed on new devices while they enroll. Uses the `/setup_experience/software` API endpoint (Fleet Premium).

## Table Usage Guide

Without a `team_id` filter the `fleetdm_setup_experience_software` table lists software for hosts with no team (`team_id = 0`) plus every team. Each team is queried once per platform (macOS, iOS, iPadOS, Windows, Linux, Android), and `platform` records which one; filter on `platform` to query just one.

## Examples

### List software installed during setup, per team

```sql+postgres
select
  team_id,
  platform,
  name,
  source
from
  fleetdm_setup_experience_software
where
  install_during_setup
order by
  team_id,
  platform,
  name;
```

```sql+sqlite
select
  team_id,
  platform,
  name,
  source
from
  fleetdm_setup_experience_software
where
  install_during_setup
order by
  team_id,
  platform,
  name;
```
