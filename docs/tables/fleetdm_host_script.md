---
title: "Steampipe Table: fleetdm_host_script - Query FleetDM Host Scripts using SQL"
description: "Allows users to query the scripts available to each FleetDM host and their last run."
---

# Table: fleetdm_host_script - Query FleetDM Host Scripts using SQL

Fleet can run saved scripts on hosts. Uses the `/hosts/{id}/scripts` API endpoint.

## Table Usage Guide

The `fleetdm_host_script` table lists scripts per host with their last execution. Without a `host_id` filter the table queries every host (up to 10 requests at a time), which can take a while on a large fleet; filter by `host_id` for a single host.

## Examples

### Find scripts whose last run failed

```sql+postgres
select
  host_id,
  name,
  last_execution ->> 'executed_at' as executed_at
from
  fleetdm_host_script
where
  last_execution ->> 'status' = 'error';
```

```sql+sqlite
select
  host_id,
  name,
  json_extract(last_execution, '$.executed_at') as executed_at
from
  fleetdm_host_script
where
  json_extract(last_execution, '$.status') = 'error';
```
