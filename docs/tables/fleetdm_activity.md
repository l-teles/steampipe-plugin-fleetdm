---
title: "Steampipe Table: fleetdm_activity - Query FleetDM Audit Log Activities using SQL"
description: "Allows users to query FleetDM audit log activities, providing insights into user actions, system events, and security-related activities within your FleetDM instance."
---

# Table: fleetdm_activity - Query FleetDM Audit Log Activities using SQL

FleetDM is an open-source device management platform that helps you manage and secure your devices. The audit log activities table captures various events such as user logins, query executions, policy creations, and other significant actions that occur within your FleetDM instance.

## Table Usage Guide

The `fleetdm_activity` table provides comprehensive insights into all audit log activities within your FleetDM instance. As a security administrator or system administrator, you can use this table to monitor user actions, track system changes, and investigate security incidents. The table helps you understand who did what, when, and with what details, enabling better security oversight and compliance management.

## Examples

### List the 100 most recent activities
Monitor recent system activities to stay informed about ongoing operations and potential security concerns. This helps in maintaining system security and identifying any unusual patterns.

```sql+postgres
select
  id,
  created_at,
  actor_full_name,
  type,
  details ->> 'public_ip' as public_ip
from
  fleetdm_activity
order by
  id desc
limit 100;
```

```sql+sqlite
select
  id,
  created_at,
  actor_full_name,
  type,
  json_extract(details, '$.public_ip') as public_ip
from
  fleetdm_activity
order by
  id desc
limit 100;
```

### Find all activities performed by a specific user
Track all actions taken by a particular user to monitor their system usage and ensure compliance with security policies.

```sql+postgres
select
  id,
  created_at,
  type,
  details
from
  fleetdm_activity
where
  actor_email = 'admin@fleetdm.com'
order by
  created_at desc;
```

```sql+sqlite
select
  id,
  created_at,
  type,
  details
from
  fleetdm_activity
where
  actor_email = 'admin@fleetdm.com'
order by
  created_at desc;
```

### List all "live_query" activities and the query executed
Analyze live query executions to understand how users are interacting with the system and what data they're accessing.

```sql+postgres
select
  created_at,
  actor_full_name,
  details ->> 'query_sql' as live_query_sql,
  details ->> 'targets_count' as live_query_targets_count
from
  fleetdm_activity
where
  type = 'live_query'
order by
  created_at desc;
```

```sql+sqlite
select
  created_at,
  actor_full_name,
  json_extract(details, '$.query_sql') as live_query_sql,
  json_extract(details, '$.targets_count') as live_query_targets_count
from
  fleetdm_activity
where
  type = 'live_query'
order by
  created_at desc;
```

### Count activities by type
Analyze the distribution of different types of activities to understand system usage patterns and identify potential areas of concern.

```sql+postgres
select
  type,
  count(*) as activity_count
from
  fleetdm_activity
group by
  type
order by
  activity_count desc;
```

```sql+sqlite
select
  type,
  count(*) as activity_count
from
  fleetdm_activity
group by
  type
order by
  activity_count desc;
```