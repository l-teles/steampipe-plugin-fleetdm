# Table: fleetdm_activity

This table lists audit log activities within your FleetDM instance. It captures events such as user logins, query executions, policy creations, and other significant actions.

## Columns

| Name                | Type        | Description                                                                                         |
| ------------------- | ----------- | --------------------------------------------------------------------------------------------------- |
| id                  | `INT`       | Unique ID of the activity.                                                                          |
| created_at          | `TIMESTAMP` | Timestamp when the activity occurred.                                                               |
| actor_full_name     | `TEXT`      | Full name of the actor who performed the activity.                                                  |
| actor_id            | `INT`       | ID of the actor (user). Null for system activities.                                                 |
| actor_email         | `TEXT`      | Email of the actor.                                                                                 |
| actor_gravatar      | `TEXT`      | Gravatar URL for the actor.                                                                         |
| type                | `TEXT`      | Type of activity (e.g., 'created_user', 'ran_live_query', 'deleted_pack').                          |
| details             | `JSONB`     | JSON object containing details specific to the activity type. The structure varies by activity type.  |
| host_id             | `INT`       | ID of the host related to this activity, if applicable.                                             |
| host_display_name   | `TEXT`      | Display name of the host related to this activity, if applicable.                                   |
| server_url          | `TEXT`      | FleetDM server URL from connection config.                                                          |

## Example Queries

**List the 100 most recent activities:**
```sql
SELECT
  id,
  created_at,
  actor_full_name,
  type,
  details ->> 'query_name' AS query_name_detail -- Example: Extract query name if present in details
FROM
  fleetdm_activity
ORDER BY
  id DESC
LIMIT 100;
```

**Find all activities performed by a specific user:**
```sql
SELECT
  id,
  created_at,
  type,
  details
FROM
  fleetdm_activity
WHERE
  actor_email = 'user@example.com' -- Replace with an actual user email
ORDER BY
  created_at DESC;
```

**List all "live_query" activities and the query executed:**
```sql
SELECT
  created_at,
  actor_full_name,
  details ->> 'query_sql' AS live_query_sql,
  details ->> 'targets_count' AS live_query_targets_count
FROM
  fleetdm_activity
WHERE
  type = 'ran_live_query' -- Or 'live_query' depending on exact type string from API
ORDER BY
  created_at DESC;
```

**Count activities by type:**
```sql
SELECT
  type,
  COUNT(*) AS activity_count
FROM
  fleetdm_activity
GROUP BY
  type
ORDER BY
  activity_count DESC;
