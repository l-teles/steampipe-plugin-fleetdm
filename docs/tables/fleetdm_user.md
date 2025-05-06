# Table: fleetdm_user

This table provides information about users registered in your FleetDM instance, including their roles and team associations.

## Columns

| Name                          | Type        | Description                                                                                         |
| ----------------------------- | ----------- | --------------------------------------------------------------------------------------------------- |
| id                            | `INT`       | Unique ID of the user.                                                                              |
| name                          | `TEXT`      | Full name of the user.                                                                              |
| email                         | `TEXT`      | Email address of the user.                                                                          |
| global_role                   | `TEXT`      | Global role of the user (e.g., 'admin', 'maintainer', 'observer'). Null if not a global role.       |
| api_only                      | `BOOLEAN`   | Indicates if the user is an API-only user.                                                          |
| sso_enabled                   | `BOOLEAN`   | Indicates if Single Sign-On is enabled for the user.                                                |
| admin_forced_password_reset   | `BOOLEAN`   | Indicates if an admin has forced a password reset for the user.                                     |
| gravatar_url                  | `TEXT`      | URL for the user's Gravatar image.                                                                  |
| created_at                    | `TIMESTAMP` | Timestamp when the user was created.                                                                |
| updated_at                    | `TIMESTAMP` | Timestamp when the user was last updated.                                                           |
| teams                         | `JSONB`     | Teams the user belongs to, including their role in each team. Array of objects: `{"id", "name", "role"}`. |
| server_url                    | `TEXT`      | FleetDM server URL from connection config.                                                          |

## Example Queries

**List all administrators:**
```sql
SELECT
  id,
  name,
  email,
  created_at
FROM
  fleetdm_user
WHERE
  global_role = 'admin'
ORDER BY
  name;
```

**Find users who are API-only:**
```sql
SELECT
  id,
  name,
  email
FROM
  fleetdm_user
WHERE
  api_only = TRUE;
```

**List users and the teams they belong to (unpacking the JSON):**
```sql
SELECT
  u.name AS user_name,
  u.email AS user_email,
  t ->> 'name' AS team_name,
  t ->> 'role' AS role_in_team
FROM
  fleetdm_user AS u,
  jsonb_array_elements(u.teams) AS t
ORDER BY
  u.name,
  team_name;
```

**Count users by global role:**
```sql
SELECT
  global_role,
  COUNT(*) AS user_count
FROM
  fleetdm_user
GROUP BY
  global_role
ORDER BY
  user_count DESC;
