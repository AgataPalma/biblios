#!/bin/sh

set -eu

state_file="/app/node_modules/.biblios-dependencies.sha256"
current_fingerprint="$({ sha256sum package.json package-lock.json; } | sha256sum | awk '{ print $1 }')"
installed_fingerprint=""

if [ -f "${state_file}" ]; then
    installed_fingerprint="$(cat "${state_file}")"
fi

if [ ! -x /app/node_modules/.bin/vite ] || [ "${installed_fingerprint}" != "${current_fingerprint}" ]; then
    echo "Frontend dependency definition changed; synchronizing node_modules with npm ci..."
    npm ci
    printf '%s\n' "${current_fingerprint}" > "${state_file}"
else
    echo "Frontend dependencies are synchronized with package-lock.json."
fi

exec "$@"
