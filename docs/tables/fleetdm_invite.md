---
title: "Steampipe Table: fleetdm_invite - Query FleetDM Invites using SQL"
description: "Allows users to query pending user invitations in FleetDM."
---

# Table: fleetdm_invite - Query FleetDM Invites using SQL

Invites are pending invitations for new Fleet users. Uses the `/invites` API endpoint; the invite token is never returned.

## Table Usage Guide

The `fleetdm_invite` table helps you review who has been invited and with which role before they accept.

## Examples

### List pending invites with admin rights

```sql+postgres
select
  email,
  name,
  global_role,
  created_at
from
  fleetdm_invite
where
  global_role = 'admin';
```

```sql+sqlite
select
  email,
  name,
  global_role,
  created_at
from
  fleetdm_invite
where
  global_role = 'admin';
```
