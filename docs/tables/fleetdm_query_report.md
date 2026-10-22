---
title: "Steampipe Table: fleetdm_query_report - Query FleetDM Query Reports using SQL"
description: "Allows users to query the results FleetDM stores for saved queries (reports)."
---

# Table: fleetdm_query_report - Query FleetDM Query Reports using SQL

Fleet stores the latest results of scheduled saved queries (reports) from each host, up to 1,000 rows per report. Uses the `/reports/{id}/report` API endpoint (`/queries/{id}/report` on older Fleet).

## Table Usage Guide

The `fleetdm_query_report` table returns one row per stored result row, with the result in the `columns` JSON. Query results can hold anything a query reads (environment variables, file contents, user names), so `columns` is NULL unless `expose_secrets = true` is set in the connection config. Without a `query_id` filter every saved query is read, global and per team (one request per team to find them, then one per query).

## Examples

### Get the stored results of one report

```sql+postgres
select
  host_name,
  last_fetched,
  columns
from
  fleetdm_query_report
where
  query_id = 1;
```

```sql+sqlite
select
  host_name,
  last_fetched,
  columns
from
  fleetdm_query_report
where
  query_id = 1;
```

### Pull one column out of the results

```sql+postgres
select
  host_name,
  columns ->> 'version' as osquery_version
from
  fleetdm_query_report
where
  query_id = 1;
```

```sql+sqlite
select
  host_name,
  json_extract(columns, '$.version') as osquery_version
from
  fleetdm_query_report
where
  query_id = 1;
```
