---
title: "Steampipe Table: fleetdm_query - Query FleetDM Saved Queries using SQL"
description: "Allows users to query FleetDM saved queries, providing insights into query configurations, scheduling, and usage patterns within your FleetDM instance."
---

# Table: fleetdm_query - Query FleetDM Saved Queries using SQL

FleetDM is an open-source device management platform that helps you manage and secure your devices. The saved queries table contains information about queries that can be run manually or scheduled to collect information from hosts in your FleetDM instance.

## Table Usage Guide

The `fleetdm_query` table provides comprehensive insights into all saved queries within your FleetDM instance. As a system administrator or security analyst, you can use this table to manage and monitor query configurations, track query usage, and ensure proper scheduling of automated queries. The table helps you understand query ownership, scheduling patterns, and performance metrics.

## Columns

| Name                          | Type        | Description                                                                                         |
| ----------------------------- | ----------- | --------------------------------------------------------------------------------------------------- |
| id                            | `INT`       | Unique ID of the saved query.                                                                       |
| name                          | `TEXT`      | Name of the saved query.                                                                            |
| query_sql                     | `TEXT`      | The SQL content of the saved query. (Mapped from API field `query`)                                 |
| description                   | `TEXT`      | Description of the saved query.                                                                     |
| team_id                       | `INT`       | ID of the team the query belongs to. Null if it's a global query.                                   |
| author_id                     | `INT`       | ID of the user who created the query.                                                               |
| author_name                   | `TEXT`      | Name of the user who created the query.                                                             |
| author_email                  | `TEXT`      | Email of the user who created the query.                                                            |
| observer_can_run              | `BOOLEAN`   | Indicates if users with the observer role can run this query.                                       |
| automations_enabled           | `BOOLEAN`   | Indicates if automations (scheduling) are enabled for this query.                                   |
| interval                      | `INT`       | Interval in seconds for scheduled execution. Null if not scheduled.                                 |
| platform                      | `TEXT`      | Target platform(s) for the query (comma-separated, or empty for all).                               |
| min_osquery_version           | `TEXT`      | Minimum osquery version required to run this query.                                                 |
| logging_type                  | `TEXT`      | Type of logging for query results (e.g., 'snapshot', 'differential', 'differential_ignore_removals'). |
| stats                         | `JSONB`     | Performance statistics for the query execution (e.g., average memory, executions, wall time).       |
| packs                         | `JSONB`     | Packs this query belongs to. (Details available on GET /api/v1/fleet/queries/{id})                  |
| created_at                    | `TIMESTAMP` | Timestamp when the query was created.                                                               |
| updated_at                    | `TIMESTAMP` | Timestamp when the query was last updated.                                                          |
| query_text_filter             | `TEXT`      | (Key Column) Search query string to filter saved queries by name or SQL. Use in `WHERE` clause.     |
| server_url                    | `TEXT`      | FleetDM server URL from connection config.                                                          |

## Examples

### List all saved queries and their authors
Get an overview of all saved queries in your FleetDM instance, including who created them and when.

```sql+postgres
select
  id,
  name,
  query_sql,
  author_name,
  created_at
from
  fleetdm_query
order by
  name;
```

```sql+sqlite
select
  id,
  name,
  query_sql,
  author_name,
  created_at
from
  fleetdm_query
order by
  name;
```

### Find queries scheduled to run
Identify all queries that have automations enabled and their scheduled intervals.

```sql+postgres
select
  id,
  name,
  query_sql,
  "interval" as schedule_interval_seconds
from
  fleetdm_query
where
  automations_enabled = true
order by
  schedule_interval_seconds;
```

```sql+sqlite
select
  id,
  name,
  query_sql,
  "interval" as schedule_interval_seconds
from
  fleetdm_query
where
  automations_enabled = true
order by
  schedule_interval_seconds;
```

### Search for queries containing "ssh"
Find all queries that contain "ssh" in their SQL content, useful for security auditing and query management.

```sql+postgres
select
  id,
  name,
  query_sql
from
  fleetdm_query
where
  query_sql like '%ssh%';
```

```sql+sqlite
select
  id,
  name,
  query_sql
from
  fleetdm_query
where
  query_sql like '%ssh%';
```

### List queries associated with a specific team
View all queries that belong to a particular team, helping with team-based query management.

```sql+postgres
select
  id,
  name,
  query_sql
from
  fleetdm_query
where
  team_id = 1;
```

```sql+sqlite
select
  id,
  name,
  query_sql
from
  fleetdm_query
where
  team_id = 1;
```