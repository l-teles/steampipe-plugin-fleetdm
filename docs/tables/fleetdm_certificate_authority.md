---
title: "Steampipe Table: fleetdm_certificate_authority - Query FleetDM Certificate Authorities using SQL"
description: "Allows users to query certificate authorities integrated with FleetDM."
---

# Table: fleetdm_certificate_authority - Query FleetDM Certificate Authorities using SQL

Fleet can request certificates for hosts from integrated certificate authorities (DigiCert, NDES, custom SCEP, Smallstep and others). Uses the `/certificate_authorities` API endpoint (Fleet Premium).

## Table Usage Guide

The `fleetdm_certificate_authority` table lists the configured certificate authorities. Only the list summary is read; credentials are never fetched.

## Examples

### List certificate authorities by type

```sql+postgres
select
  type,
  name
from
  fleetdm_certificate_authority
order by
  type,
  name;
```

```sql+sqlite
select
  type,
  name
from
  fleetdm_certificate_authority
order by
  type,
  name;
```
