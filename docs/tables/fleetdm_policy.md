# Table: fleetdm_policy

This table lists the policies defined in FleetDM. Policies are osquery queries that determine if a host is passing or failing a specific security or configuration guideline. This table specifically queries global policies (from `/api/v1/fleet/global/policies`).

## Columns

| Name                          | Type        | Description                                                                                         |
| ----------------------------- | ----------- | --------------------------------------------------------------------------------------------------- |
| id                            | `INT`       | Unique ID of the policy.                                                                            |
| name                          | `TEXT`      | Name of the policy.                                                                                 |
| query_text                    | `TEXT`      | The osquery query that defines the policy. (Mapped from API field `query`)                          |
| description                   | `TEXT`      | Description of the policy.                                                                          |
| platform                      | `TEXT`      | Target platform for the policy (e.g., 'darwin', 'windows', 'linux', or empty for all).              |
| team_id                       | `INT`       | ID of the team the policy belongs to. Null if it's a global policy (which this table primarily lists). |
| passing_host_count            | `INT`       | Number of hosts currently passing this policy.                                                      |
| failing_host_count            | `INT`       | Number of hosts currently failing this policy.                                                      |
| resolution                    | `TEXT`      | Resolution steps or instructions for hosts failing this policy.                                     |
| author_id                     | `INT`       | ID of the user who created the policy.                                                              |
| author_name                   | `TEXT`      | Name of the user who created the policy.                                                            |
| author_email                  | `TEXT`      | Email of the user who created the policy.                                                           |
| critical                      | `BOOLEAN`   | Whether the policy is marked as critical.                                                           |
| calendar_events_enabled       | `BOOLEAN`   | Whether calendar events are enabled for this policy.                                                |
| created_at                    | `TIMESTAMP` | Timestamp when the policy was created.                                                              |
| updated_at                    | `TIMESTAMP` | Timestamp when the policy was last updated.                                                         |
| filter_search_query           | `TEXT`      | (Key Column) Search query string to filter policies by name or query text. Use in `WHERE` clause.   |
| server_url                    | `TEXT`      | FleetDM server URL from connection config.                                                          |

## Example Queries

**List all global policies and their pass/fail counts:**
```sql
SELECT
  id,
  name,
  platform,
  passing_host_count,
  failing_host_count
FROM
  fleetdm_policy
ORDER BY
  name;
```

**Find critical policies:**
```sql
SELECT
  id,
  name,
  query_text,
  failing_host_count
FROM
  fleetdm_policy
WHERE
  critical = TRUE
ORDER BY
  failing_host_count DESC;
```

**Search for policies related to "firewall":**
```sql
SELECT
  id,
  name,
  query_text
FROM
  fleetdm_policy
WHERE
  filter_search_query = 'firewall';
```

**Count policies by platform:**
```sql
SELECT
  platform,
  COUNT(*) AS policy_count
FROM
  fleetdm_policy
GROUP BY
  platform
ORDER BY
  policy_count DESC;
