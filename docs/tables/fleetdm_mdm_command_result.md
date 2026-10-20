---
title: "Steampipe Table: fleetdm_mdm_command_result - Query FleetDM MDM Command Results using SQL"
description: "Allows users to query the results of an MDM command on each targeted host."
---

# Table: fleetdm_mdm_command_result - Query FleetDM MDM Command Results using SQL

Each MDM command produces a result per targeted host. Uses the `/commands/results` API endpoint.

## Table Usage Guide

The `fleetdm_mdm_command_result` table requires `command_uuid`. The decoded `payload` and `result` can contain sensitive values (for example a lock PIN), so they are NULL unless `expose_secrets = true` is set in the connection config.

## Examples

### Get the results of a command

```sql+postgres
select
  hostname,
  status,
  updated_at
from
  fleetdm_mdm_command_result
where
  command_uuid = '00000000-0000-0000-0000-000000000000';
```

```sql+sqlite
select
  hostname,
  status,
  updated_at
from
  fleetdm_mdm_command_result
where
  command_uuid = '00000000-0000-0000-0000-000000000000';
```
