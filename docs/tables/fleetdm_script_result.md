---
title: "Steampipe Table: fleetdm_script_result - Query FleetDM Script Results using SQL"
description: "Allows users to query the result of a FleetDM script execution."
---

# Table: fleetdm_script_result - Query FleetDM Script Results using SQL

Each script run on a host has an execution ID and a result. Uses the `/scripts/results/{execution_id}` API endpoint.

## Table Usage Guide

The `fleetdm_script_result` table requires `execution_id` (the `script_execution_id` column of `fleetdm_host_activity`). `script_contents` and `output` can contain anything the script handles, so they are NULL unless `expose_secrets = true` is set in the connection config.

## Examples

### Get a script result

```sql+postgres
select
  host_id,
  exit_code,
  runtime,
  message
from
  fleetdm_script_result
where
  execution_id = '00000000-0000-0000-0000-000000000000';
```

```sql+sqlite
select
  host_id,
  exit_code,
  runtime,
  message
from
  fleetdm_script_result
where
  execution_id = '00000000-0000-0000-0000-000000000000';
```
