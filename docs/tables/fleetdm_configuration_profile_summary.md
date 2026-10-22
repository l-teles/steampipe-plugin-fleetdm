---
title: "Steampipe Table: fleetdm_configuration_profile_summary - Query FleetDM Configuration Profile Summary using SQL"
description: "Allows users to query FleetDM's configuration profile delivery status counts."
---

# Table: fleetdm_configuration_profile_summary - Query FleetDM Configuration Profile Summary using SQL

Fleet delivers configuration profiles (OS settings) through MDM and verifies they are installed. Uses the `/configuration_profiles/summary` API endpoint.

## Table Usage Guide

The `fleetdm_configuration_profile_summary` table returns one row. Filter with `team_id` (0 for hosts with no team). It returns no rows when MDM is off.

## Examples

### Check profile delivery for a team

```sql+postgres
select
  verified,
  verifying,
  pending,
  failed
from
  fleetdm_configuration_profile_summary
where
  team_id = 1;
```

```sql+sqlite
select
  verified,
  verifying,
  pending,
  failed
from
  fleetdm_configuration_profile_summary
where
  team_id = 1;
```
