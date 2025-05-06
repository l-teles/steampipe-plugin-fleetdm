# FleetDM Plugin Documentation

The FleetDM plugin for Steampipe allows you to query and analyze data from your [FleetDM](https://fleetdm.com/) instance using SQL.
FleetDM is an open-source device management platform that collects a wealth of information about your hosts (laptops, servers, etc.) using osquery. With this Steampipe plugin, you can access detailed information about your hosts, software inventory, users, teams, policies, queries, packs, labels, and audit activities directly from your terminal or by connecting to the Steampipe database.

This empowers you to perform security audits, compliance checks, asset inventory, and operational monitoring with the flexibility and power of SQL.

## Overview

The FleetDM plugin maps the various resources and data points within FleetDM to SQL tables. This allows you to:

* **Query Host Information:** Get detailed hardware, operating system, network, and configuration data for all your managed devices.
* **Manage Software Inventory:** List installed software, versions, and identify hosts running specific applications.
* **Monitor Security Policies:** Check policy compliance status across your fleet.
* **Audit Activities:** Review user actions, system events, and live query executions.
* **Cross-Reference Data:** Join FleetDM data with other security and infrastructure data sources using Steampipe's multi-plugin capabilities.

## Quick Start

Before you begin, ensure you have [Steampipe](https://steampipe.io/downloads) installed and have [configured the FleetDM plugin](https://github.com/l-teles/steampipe-plugin-fleetdm/blob/main/README.md#configuration) with your FleetDM server URL and API token.

Once configured, you can start querying your FleetDM data. For example, to list all your hosts:

```sql
SELECT
  id,
  hostname,
  platform,
  os_version,
  status,
  team_name
FROM
  fleetdm_host
ORDER BY
  hostname;
```

## Examples

Here are a few examples of what you can do with the FleetDM plugin. For more detailed examples, please see the documentation for each table.

**Find all Windows hosts not in any team:**
```sql
SELECT
  hostname,
  os_version,
  public_ip
FROM
  fleetdm_host
WHERE
  platform = 'windows'
  AND team_id IS NULL;
```

**Identify users who are administrators:**
```sql
SELECT
  name,
  email
FROM
  fleetdm_user
WHERE
  global_role = 'admin';
```

**List policies that are failing on more than 5 hosts:**
```sql
SELECT
  name,
  query_text,
  failing_host_count
FROM
  fleetdm_policy
WHERE
  failing_host_count > 5
ORDER BY
  failing_host_count DESC;
```

## Tables

Click on a table name to see an overview, list of columns, and example queries.

| Table                                       | Description                                                                                                |
| ------------------------------------------- | ---------------------------------------------------------------------------------------------------------- |
| [fleetdm_activity](./tables/fleetdm_activity.md) | Audit log activities within your FleetDM instance, capturing events like logins and query executions.      |
| [fleetdm_host](./tables/fleetdm_host.md)         | Detailed information about each host (laptops, servers, etc.) managed by FleetDM.                          |
| [fleetdm_label](./tables/fleetdm_label.md)       | Labels defined in FleetDM used for grouping hosts, either manually or dynamically.                         |
| [fleetdm_pack](./tables/fleetdm_pack.md)         | Query packs, which are collections of queries scheduled to run against targeted hosts or labels.           |
| [fleetdm_policy](./tables/fleetdm_policy.md)     | Policies defined in FleetDM, which are osquery queries determining host compliance.                        |
| [fleetdm_query](./tables/fleetdm_query.md)       | Saved queries that can be run manually or scheduled to collect information from hosts.                     |
| [fleetdm_software](./tables/fleetdm_software.md) | Software items inventoried by FleetDM across all managed hosts, including version and vulnerability data.    |
| [fleetdm_team](./tables/fleetdm_team.md)         | Teams configured in FleetDM, used for grouping users and managing host configurations and policies.        |
| [fleetdm_user](./tables/fleetdm_user.md)         | Users registered in your FleetDM instance, including their roles and team associations.                    |

## Getting Started

For full installation and configuration instructions, please refer to the main [README.md](https://github.com/l-teles/steampipe-plugin-fleetdm/blob/main/README.md) for this plugin.
Ensure you have [Steampipe](https://steampipe.io/downloads) installed and have [configured the FleetDM plugin connection](https://github.com/l-teles/steampipe-plugin-fleetdm/blob/main/README.md#configuration) with your FleetDM server URL and API token.
