---
title: "Steampipe Table: fleetdm_host_activity - Query FleetDM Host Activities using SQL"
description: "Allows users to query the past activities on each FleetDM host."
---

# Table: fleetdm_host_activity - Query FleetDM Host Activities using SQL

Fleet records actions taken on hosts, such as scripts run, software installed, MDM enrollment and lock or wipe. Uses the `/hosts/{id}/activities` API endpoint.

## Table Usage Guide

The `fleetdm_host_activity` table lists past activities per host. Without a `host_id` filter the table queries every host (up to 10 requests at a time); filter by `host_id` for a single host. `script_execution_id`, `install_uuid` and `command_uuid` link activities to `fleetdm_script_result`, `fleetdm_software_install_result` and `fleetdm_mdm_command_result`.

## Examples

### List script runs on a host with their exit code

Join on the plain `script_execution_id` column, which `fleetdm_script_result` needs as its required filter. Materializing the filtered activities first keeps Postgres from planning the result table without that filter.

```sql+postgres
with runs as materialized (
  select
    created_at,
    actor_email,
    script_execution_id
  from
    fleetdm_host_activity
  where
    host_id = 1
    and type = 'ran_script'
)
select
  runs.created_at,
  runs.actor_email,
  r.exit_code
from
  runs
  join fleetdm_script_result r on r.execution_id = runs.script_execution_id;
```

```sql+sqlite
select
  a.created_at,
  a.actor_email,
  r.exit_code
from
  fleetdm_host_activity a
  join fleetdm_script_result r on r.execution_id = a.script_execution_id
where
  a.host_id = 1
  and a.type = 'ran_script';
```
