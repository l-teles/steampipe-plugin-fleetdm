---
title: "Steampipe Table: fleetdm_app_store_app - Query FleetDM Apple App Store Apps using SQL"
description: "Allows users to query FleetDM Apple App Store (VPP) apps, providing insights into available App Store apps across teams for deployment and management."
---

# Table: fleetdm_app_store_app - Query FleetDM Apple App Store Apps using SQL

FleetDM is an open-source device management platform that helps you manage and secure your devices. The App Store apps table provides information about Apple App Store apps (VPP) that are available for installation on your managed devices. Since the API requires a team ID, the plugin automatically discovers all teams and retrieves App Store apps for each one. Uses the `/software/app_store_apps` API endpoint.

## Table Usage Guide

The `fleetdm_app_store_app` table provides insights into Apple App Store (VPP) apps available in your FleetDM instance. As a system administrator or IT manager, you can use this table to audit which App Store apps are available for deployment, track app versions, and manage app distribution across teams. You can filter by `team_id` in the WHERE clause to query a specific team, or omit it to automatically discover and query all teams.

## Examples

### List all App Store apps across all teams

Get an overview of all Apple App Store apps available across all teams in your FleetDM instance.

```sql+postgres
select
  name,
  app_store_id,
  platform,
  latest_version,
  team_id,
  team_name
from
  fleetdm_app_store_app
order by
  name, platform;
```

```sql+sqlite
select
  name,
  app_store_id,
  platform,
  latest_version,
  team_id,
  team_name
from
  fleetdm_app_store_app
order by
  name, platform;
```

### List App Store apps for a specific team

View the App Store apps available for a particular team.

```sql+postgres
select
  name,
  app_store_id,
  platform,
  latest_version,
  bundle_identifier
from
  fleetdm_app_store_app
where
  team_id = 1
order by
  name;
```

```sql+sqlite
select
  name,
  app_store_id,
  platform,
  latest_version,
  bundle_identifier
from
  fleetdm_app_store_app
where
  team_id = 1
order by
  name;
```

### Find App Store apps by platform

Identify App Store apps available for a specific platform (e.g., macOS, iOS, iPadOS).

```sql+postgres
select
  name,
  app_store_id,
  latest_version,
  bundle_identifier,
  team_name
from
  fleetdm_app_store_app
where
  platform = 'darwin'
order by
  name;
```

```sql+sqlite
select
  name,
  app_store_id,
  latest_version,
  bundle_identifier,
  team_name
from
  fleetdm_app_store_app
where
  platform = 'darwin'
order by
  name;
```

### List self-service App Store apps

Identify App Store apps that are configured for self-service installation by end users.

```sql+postgres
select
  name,
  platform,
  latest_version,
  team_name
from
  fleetdm_app_store_app
where
  self_service = true
order by
  name;
```

```sql+sqlite
select
  name,
  platform,
  latest_version,
  team_name
from
  fleetdm_app_store_app
where
  self_service = 1
order by
  name;
```

### Count App Store apps per team

Get a summary of how many App Store apps are available per team.

```sql+postgres
select
  team_id,
  team_name,
  count(*) as app_count
from
  fleetdm_app_store_app
group by
  team_id, team_name
order by
  app_count desc;
```

```sql+sqlite
select
  team_id,
  team_name,
  count(*) as app_count
from
  fleetdm_app_store_app
group by
  team_id, team_name
order by
  app_count desc;
```
