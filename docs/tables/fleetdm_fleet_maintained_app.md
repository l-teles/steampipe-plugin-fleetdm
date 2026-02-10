---
title: "Steampipe Table: fleetdm_fleet_maintained_app - Query FleetDM Fleet-Maintained Apps using SQL"
description: "Allows users to query FleetDM Fleet-maintained apps, providing insights into pre-packaged software installers maintained by Fleet for easy deployment."
---

# Table: fleetdm_fleet_maintained_app - Query FleetDM Fleet-Maintained Apps using SQL

FleetDM is an open-source device management platform that helps you manage and secure your devices. Fleet-maintained apps are pre-packaged software installers that are maintained and updated by Fleet, making it easy to deploy and manage common applications across your fleet. Uses the `/software/fleet_maintained_apps` API endpoint.

## Table Usage Guide

The `fleetdm_fleet_maintained_app` table provides insights into the catalog of Fleet-maintained apps available for deployment. As a system administrator, you can use this table to discover available pre-packaged applications, check their current versions, and identify which apps have already been added to specific teams. This table is especially useful for software deployment planning and standardization.

## Examples

### List all Fleet-maintained apps

Get an overview of all available Fleet-maintained apps in the catalog.

```sql+postgres
select
  id,
  name,
  platform,
  version,
  slug
from
  fleetdm_fleet_maintained_app
order by
  name;
```

```sql+sqlite
select
  id,
  name,
  platform,
  version,
  slug
from
  fleetdm_fleet_maintained_app
order by
  name;
```

### List Fleet-maintained apps by platform

Find all available Fleet-maintained apps for a specific platform.

```sql+postgres
select
  name,
  version,
  slug,
  categories
from
  fleetdm_fleet_maintained_app
where
  platform = 'darwin'
order by
  name;
```

```sql+sqlite
select
  name,
  version,
  slug,
  categories
from
  fleetdm_fleet_maintained_app
where
  platform = 'darwin'
order by
  name;
```

### Find Fleet-maintained apps already added to a team

Identify which Fleet-maintained apps have already been added to a specific team by checking for a non-null software_title_id.

```sql+postgres
select
  name,
  version,
  platform,
  software_title_id
from
  fleetdm_fleet_maintained_app
where
  team_id = 1
  and software_title_id is not null
order by
  name;
```

```sql+sqlite
select
  name,
  version,
  platform,
  software_title_id
from
  fleetdm_fleet_maintained_app
where
  team_id = 1
  and software_title_id is not null
order by
  name;
```

### List Fleet-maintained apps with their categories

Explore the categories of Fleet-maintained apps to understand the types of software available.

```sql+postgres
select
  name,
  platform,
  version,
  c as category
from
  fleetdm_fleet_maintained_app,
  jsonb_array_elements_text(categories) as c
order by
  category, name;
```

```sql+sqlite
select
  name,
  platform,
  version,
  c.value as category
from
  fleetdm_fleet_maintained_app,
  json_each(categories) as c
order by
  c.value, name;
```
