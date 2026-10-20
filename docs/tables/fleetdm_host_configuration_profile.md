---
title: "Steampipe Table: fleetdm_host_configuration_profile - Query FleetDM Host Configuration Profiles using SQL"
description: "Allows users to query the configuration profiles assigned to each FleetDM host."
---

# Table: fleetdm_host_configuration_profile - Query FleetDM Host Configuration Profiles using SQL

Fleet assigns configuration profiles (OS settings) to hosts through MDM. Uses the `/hosts/{id}/configuration_profiles` API endpoint. Per-host delivery status is in `fleetdm_host_detail.mdm`.

## Table Usage Guide

The `fleetdm_host_configuration_profile` table lists profiles per host. Without a `host_id` filter the table queries every host (up to 10 requests at a time), which can take a while on a large fleet; filter by `host_id` for a single host.

## Examples

### List the profiles on one host

```sql+postgres
select
  name,
  platform,
  identifier
from
  fleetdm_host_configuration_profile
where
  host_id = 1;
```

```sql+sqlite
select
  name,
  platform,
  identifier
from
  fleetdm_host_configuration_profile
where
  host_id = 1;
```
