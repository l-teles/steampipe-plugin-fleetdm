---
title: "Steampipe Table: fleetdm_user - Query FleetDM Users using SQL"
description: "Allows users to query FleetDM user accounts, providing insights into user roles, team memberships, and access patterns within your FleetDM instance."
---

# Table: fleetdm_user - Query FleetDM Users using SQL

FleetDM is an open-source device management platform that helps you manage and secure your devices. The users table provides comprehensive information about user accounts, including their roles, team associations, and authentication settings.

## Table Usage Guide

The `fleetdm_user` table provides detailed insights into user management within your FleetDM instance. As a system administrator, you can use this table to monitor user roles, track team memberships, and manage access control. The table helps you understand user distribution, role assignments, and authentication configurations.

**Note:** This table supports the following optional key columns for server-side API filtering:

- `query` — Search users by name or email.
- `team_id` — Filter users by team (Fleet Premium).

Using these key columns in your `WHERE` clause pushes the filtering to the FleetDM API, reducing data transfer and improving query performance.

## Examples

### List all administrators

Identify all users with administrative privileges to ensure proper access control and security oversight.

```sql+postgres
select
  id,
  name,
  email,
  created_at
from
  fleetdm_user
where
  global_role = 'admin'
order by
  name;
```

```sql+sqlite
select
  id,
  name,
  email,
  created_at
from
  fleetdm_user
where
  global_role = 'admin'
order by
  name;
```

### Find users who are API-only

Identify users that are configured for API access only, useful for service account management.

```sql+postgres
select
  id,
  name,
  email
from
  fleetdm_user
where
  api_only = true;
```

```sql+sqlite
select
  id,
  name,
  email
from
  fleetdm_user
where
  api_only = true;
```

### List users and the teams they belong to

Get a comprehensive view of user team memberships and their roles within each team.

```sql+postgres
select
  u.name as user_name,
  u.email as user_email,
  t ->> 'name' as team_name,
  t ->> 'role' as role_in_team
from
  fleetdm_user as u,
  jsonb_array_elements(u.teams) as t
order by
  u.name,
  team_name;
```

```sql+sqlite
select
  u.name as user_name,
  u.email as user_email,
  json_extract(t.value, '$.name') as team_name,
  json_extract(t.value, '$.role') as role_in_team
from
  fleetdm_user as u,
  json_each(u.teams) as t
order by
  u.name,
  team_name;
```

### Search users by name or email

Use the `query` key column to search users server-side by name or email.

```sql+postgres
select
  id,
  name,
  email,
  global_role
from
  fleetdm_user
where
  query = 'admin';
```

```sql+sqlite
select
  id,
  name,
  email,
  global_role
from
  fleetdm_user
where
  query = 'admin';
```

### List users belonging to a specific team

Use the `team_id` key column to filter users by team membership (Fleet Premium).

```sql+postgres
select
  id,
  name,
  email
from
  fleetdm_user
where
  team_id = 3;
```

```sql+sqlite
select
  id,
  name,
  email
from
  fleetdm_user
where
  team_id = 3;
```

### Count users by global role

Analyze the distribution of user roles across your FleetDM instance to ensure proper role allocation.

```sql+postgres
select
  global_role,
  count(*) as user_count
from
  fleetdm_user
group by
  global_role
order by
  user_count desc;
```

```sql+sqlite
select
  global_role,
  count(*) as user_count
from
  fleetdm_user
group by
  global_role
order by
  user_count desc;
```
