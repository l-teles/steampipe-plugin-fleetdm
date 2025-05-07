# Table: fleetdm_query

This table lists the saved queries in your FleetDM instance. Saved queries can be run manually or scheduled to collect information from hosts.

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

## Example Queries

**List all saved queries and their authors:**
```sql
SELECT
  id,
  name,
  query_sql,
  author_name,
  created_at
FROM
  fleetdm_query
ORDER BY
  name;
```

**Find queries scheduled to run (automations enabled):**
```sql
SELECT
  id,
  name,
  query_sql,
  "interval" AS schedule_interval_seconds
FROM
  fleetdm_query
WHERE
  automations_enabled = TRUE
ORDER BY
  schedule_interval_seconds;
```

**Search for queries containing "ssh" as part of the search:**
```sql
SELECT
  id,
  name,
  query_sql
FROM
  fleetdm_query
WHERE
  query_sql LIKE '%ssh%';
```

**List queries associated with a specific team:**
```sql
SELECT
  id,
  name,
  query_sql
FROM
  fleetdm_query
WHERE
  team_id = 1; -- Replace with an existing team ID
```