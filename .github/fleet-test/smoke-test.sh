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
  esac
}

plugin_dir="${INSTALL_DIR}/plugins/local/fleetdm"
mkdir -p "$plugin_dir" "${INSTALL_DIR}/config"
go build -o "${plugin_dir}/steampipe-plugin-fleetdm.plugin" .

cat > "${INSTALL_DIR}/config/${CONN}.spc" <<EOF
connection "${CONN}" {
  plugin = "local/fleetdm"
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

# A new connection's schema is imported asynchronously; wait for it.
tables=""
for _ in $(seq 1 20); do
  tables=$(steampipe query --output csv --header=false \
    "select table_name from information_schema.tables where table_schema = '${CONN}' order by table_name")
  [ -n "$tables" ] && break
  sleep 3
done
if [ -z "$tables" ]; then
  echo "ERROR: plugin exposed no tables; it failed to load." >&2
  exit 1
fi

failed=0
for t in $tables; do
  sql="select count(*) from ${CONN}.${t} $(extra_where "$t")"
  # steampipe can exit 0 on a failed query, so also inspect the output.
  if out=$(steampipe query --output csv --header=false "$sql" 2>&1) && ! grep -qE '^(Error|Warning)' <<<"$out"; then
    printf '%-32s ok (%s rows)\n' "$t" "$out"
  else
    printf '%-32s FAILED: %s\n' "$t" "$out"
    failed=1
  fi
done

exit "$failed"
