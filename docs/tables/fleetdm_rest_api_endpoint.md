---
title: "Steampipe Table: fleetdm_rest_api_endpoint - Query FleetDM REST API Endpoints using SQL"
description: "Allows users to query the FleetDM REST API endpoints API-only users can be limited to."
---

# Table: fleetdm_rest_api_endpoint - Query FleetDM REST API Endpoints using SQL

Fleet publishes a catalog of its REST API endpoints, used to restrict what API-only users can call. Uses the `/rest_api` API endpoint (Fleet Premium, 4.90+).

## Table Usage Guide

The `fleetdm_rest_api_endpoint` table returns no rows on older Fleet versions.

## Examples

### List deprecated endpoints

```sql+postgres
select
  method,
  path,
  display_name
from
  fleetdm_rest_api_endpoint
where
  deprecated;
```

```sql+sqlite
select
  method,
  path,
  display_name
from
  fleetdm_rest_api_endpoint
where
  deprecated;
```
