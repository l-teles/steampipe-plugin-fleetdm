#!/usr/bin/env bash
# smoke-test.sh – Build the plugin, load it into Steampipe and query every
# table it exposes against a live Fleet. Catches runtime breakage (SDK/cache
# regressions, API envelope changes) that unit tests cannot see.
#
# Requires steampipe on PATH and FLEETDM_URL / FLEETDM_API_TOKEN in the
# environment (from setup-fleet.sh). The connection config holds no
# credentials: the plugin reads them from the environment.

set -euo pipefail

INSTALL_DIR="${STEAMPIPE_INSTALL_DIR:-$HOME/.steampipe}"
CONN="fleetdm_smoke"
: "${FLEETDM_URL:?FLEETDM_URL must be set}"
: "${FLEETDM_API_TOKEN:?FLEETDM_API_TOKEN must be set}"

# Tables with required key columns get a qual; a missing row is not an error.
extra_where() {
  case "$1" in
    fleetdm_host_detail) echo "where id = 0" ;;
    fleetdm_mdm_command) echo "where host_identifier = 'smoke-no-such-host'" ;;
    fleetdm_mdm_command_result) echo "where command_uuid = 'smoke-no-such-command'" ;;
    fleetdm_script_result) echo "where execution_id = 'smoke-no-such-execution'" ;;
    fleetdm_software_install_result) echo "where install_uuid = 'smoke-no-such-install'" ;;
  esac
}

plugin_dir="${INSTALL_DIR}/plugins/local/fleetdm"
mkdir -p "$plugin_dir" "${INSTALL_DIR}/config"
go build -o "${plugin_dir}/steampipe-plugin-fleetdm.plugin" .

cat > "${INSTALL_DIR}/config/${CONN}.spc" <<EOF
connection "${CONN}" {
  plugin = "local/fleetdm"
  # Throwaway test instance: exercise the tables and columns gated on secrets.
  expose_secrets = true
}
EOF

# Each standalone `steampipe query` shuts its database down when it ends,
# which can abort the background load of a new connection (seen on Linux:
# "RefreshConnections failed ... closed pool"). Keep a service up instead.
# Reuse a running service; otherwise start one (loopback only) and stop it on
# exit. "Not running" text differs on a fresh install, so test for running.
if ! steampipe service status 2>&1 | grep -q 'Steampipe service is running'; then
  if ! out=$(steampipe service start --database-listen local 2>&1); then
    echo "ERROR: could not start the Steampipe service: $out" >&2
    exit 1
  fi
  trap 'steampipe service stop >/dev/null 2>&1 || true' EXIT
fi

# Steampipe imports a new or rebuilt plugin's schema asynchronously; wait
# until every table registered in plugin.go is visible, so none is skipped.
expected=$(grep -cE '^[[:space:]]+"fleetdm_[a-z_]+":[[:space:]]+tableFleetdm' fleetdm/plugin.go)
tables=""
for _ in $(seq 1 40); do
  tables=$(steampipe query --output csv --header=false \
    "select table_name from information_schema.tables where table_schema = '${CONN}' order by table_name")
  [ "$(grep -c . <<<"$tables")" -eq "$expected" ] && break
  sleep 3
done
if [ "$(grep -c . <<<"$tables")" -ne "$expected" ]; then
  echo "ERROR: expected ${expected} tables, Steampipe exposes $(grep -c . <<<"$tables"); the plugin failed to load." >&2
  exit 1
fi

# Enroll one fake host (as an agent would) so the per-host tables exercise
# their fetch path. Reuses the existing global enroll secret.
secret=$(curl -sf "${FLEETDM_URL}/api/v1/fleet/spec/enroll_secret" \
  -H "Authorization: Bearer ${FLEETDM_API_TOKEN}" | jq -r '.spec.secrets[0].secret // empty')
if [ -z "$secret" ]; then
  echo "ERROR: no global enroll secret to enroll the smoke-test host with." >&2
  exit 1
fi
curl -sf -o /dev/null -X POST "${FLEETDM_URL}/api/osquery/enroll" -H "Content-Type: application/json" \
  -d "{\"enroll_secret\": \"${secret}\", \"host_identifier\": \"smoke-host\", \"host_details\": {
        \"system_info\": {\"hostname\": \"smoke-host\", \"uuid\": \"00000000-0000-0000-0000-00000000a0a0\"},
        \"os_version\": {\"platform\": \"darwin\", \"name\": \"macOS\", \"version\": \"15.0\"}}}"

# Seed objects that a fresh Fleet lacks, so lookups by ID have something to
# find. These endpoints exist on every supported version.
auth=(-H "Authorization: Bearer ${FLEETDM_API_TOKEN}" -H "Content-Type: application/json")
curl -sf -o /dev/null "${auth[@]}" -X POST "${FLEETDM_URL}/api/v1/fleet/queries" \
  -d "{\"name\": \"smoke-query-$(date +%s)\", \"query\": \"SELECT 1;\"}"
curl -sf -o /dev/null "${auth[@]}" -X POST "${FLEETDM_URL}/api/v1/fleet/global/policies" \
  -d "{\"name\": \"smoke-policy-$(date +%s)\", \"query\": \"SELECT 1;\"}"
curl -sf -o /dev/null "${auth[@]}" -X POST "${FLEETDM_URL}/api/v1/fleet/packs" \
  -d "{\"name\": \"smoke-pack-$(date +%s)\"}"

failed=0
# run_check LABEL SQL: steampipe can exit 0 on a failed query, so also
# inspect the output.
run_check() {
  local out
  if out=$(steampipe query --output csv --header=false "$2" 2>&1) && ! grep -qE '^(Error|Warning)' <<<"$out"; then
    printf '%-48s ok (%s rows)\n' "$1" "$out"
  else
    printf '%-48s FAILED: %s\n' "$1" "$out"
    failed=1
  fi
}

# select * on a few rows also runs every column's hydrate and transform,
# which count(*) skips.
run_select_all() {
  local out
  if out=$(steampipe query --output csv "$2" 2>&1) && ! grep -qE '^(Error|Warning)' <<<"$out"; then
    printf '%-48s ok (%s rows, all columns)\n' "$1" "$(($(grep -c . <<<"$out") - 1))"
  else
    printf '%-48s FAILED: %s\n' "$1" "$out"
    failed=1
  fi
}

for t in $tables; do
  run_check "$t" "select count(*) from ${CONN}.${t} $(extra_where "$t")"
  # host_detail's id = 0 qual would skip the per-host detail fetch; the
  # enrolled smoke host gives it a real row instead.
  where="$(extra_where "$t")"
  [ "$t" = "fleetdm_host_detail" ] && where=""
  run_select_all "$t (select *)" "select * from ${CONN}.${t} ${where} limit 5"
done

# Lookup by ID takes a separate code path (a direct request); check it returns
# exactly the row the list has.
id_column() {
  case "$1" in
    fleetdm_policy | fleetdm_query | fleetdm_label | fleetdm_team | fleetdm_user | fleetdm_script | fleetdm_software_title | fleetdm_pack | fleetdm_carve) echo "id" ;;
    fleetdm_os_version) echo "os_version_id" ;;
    fleetdm_vulnerability) echo "cve" ;;
    fleetdm_configuration_profile) echo "profile_uuid" ;;
  esac
}
for t in $tables; do
  col=$(id_column "$t")
  [ -n "$col" ] || continue
  id=$(steampipe query --output csv --header=false "select ${col} from ${CONN}.${t} order by ${col} limit 1" 2>&1)
  if grep -qE '^(Error|Warning)' <<<"$id"; then
    printf '%-48s FAILED: %s\n' "$t (by ${col})" "$id"
    failed=1
    continue
  fi
  if [ -z "$id" ]; then
    printf '%-48s skipped (no rows to look up)\n' "$t (by ${col})"
    continue
  fi
  out=$(steampipe query --output csv --header=false "select ${col} from ${CONN}.${t} where ${col} = '${id}'" 2>&1)
  if [ "$out" = "$id" ]; then
    printf '%-48s ok (1 rows)\n' "$t (by ${col})"
  else
    printf '%-48s FAILED: want one row %s, got: %s\n' "$t (by ${col})" "$id" "$out"
    failed=1
  fi
done

# Team filters are renamed across Fleet versions (team_id -> fleet_id), so
# exercise every table that takes one against a real team.
team_id=$(curl -sf -X POST "${FLEETDM_URL}/api/v1/fleet/teams" \
  -H "Authorization: Bearer ${FLEETDM_API_TOKEN}" -H "Content-Type: application/json" \
  -d "{\"name\": \"smoke-$(date +%s)\"}" | jq -r '.team.id // .fleet.id')
if ! [[ "$team_id" =~ ^[0-9]+$ ]]; then
  echo "ERROR: could not create a smoke-test team (got '${team_id}')." >&2
  exit 1
fi
team_tables=$(steampipe query --output csv --header=false \
  "select table_name || ':' || data_type from information_schema.columns where table_schema = '${CONN}' and column_name = 'team_id' order by table_name")
for entry in $team_tables; do
  t="${entry%%:*}"
  value="${team_id}"
  [ "${entry#*:}" = "text" ] && value="'${team_id}'"
  # No extra_where here: the team filter must reach each table's list call.
  run_check "$t (team_id)" "select count(*) from ${CONN}.${t} where team_id = ${value}"
done

exit "$failed"
