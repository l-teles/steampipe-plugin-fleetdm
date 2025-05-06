# Table: fleetdm_team

This table lists the teams configured in your FleetDM instance. Teams are used to group users and can have their own set of hosts, policies, and configurations.

## Columns

| Name                | Type        | Description                                                                                         |
| ------------------- | ----------- | --------------------------------------------------------------------------------------------------- |
| id                  | `INT`       | Unique ID of the team.                                                                              |
| name                | `TEXT`      | Name of the team.                                                                                   |
| description         | `TEXT`      | Description of the team.                                                                            |
| user_count          | `INT`       | Number of users in the team.                                                                        |
| host_count          | `INT`       | Number of hosts assigned to the team.                                                               |
| created_at          | `TIMESTAMP` | Timestamp when the team was created.                                                                |
| agent_options       | `JSONB`     | Agent options configured for this team (e.g., osquery configurations).                              |
| secrets             | `JSONB`     | Enrollment secrets associated with the team. Array of objects. (Details available on GET /teams/{id}) |
| users               | `JSONB`     | Users belonging to this team and their roles. (Details available on GET /teams/{id})                |
| server_url          | `TEXT`      | FleetDM server URL from connection config.                                                          |

## Example Queries

**List all teams and their user/host counts:**
```sql
SELECT
  id,
  name,
  description,
  user_count,
  host_count,
  created_at
FROM
  fleetdm_team
ORDER BY
  name;
```

**Find teams with more than 100 hosts:**
```sql
SELECT
  id,
  name,
  host_count
FROM
  fleetdm_team
WHERE
  host_count > 100
ORDER BY
  host_count DESC;
```

**Get agent options for a specific team:**
```sql
SELECT
  name,
  agent_options -> 'config' -> 'decorators' ->> 'load' AS agent_config_decorators_load
FROM
  fleetdm_team
WHERE
  name = 'Engineering'; -- Replace with an actual team name
```

**List users associated with a specific team (if `users` field is populated by GET):**
```sql
SELECT
  t.name AS team_name,
  u ->> 'name' AS user_name,
  u ->> 'email' AS user_email,
  u ->> 'role' AS role_in_team
FROM
  fleetdm_team AS t,
  jsonb_array_elements(t.users) AS u
WHERE
  t.id = 1; -- Replace with an actual team ID
```