#!/usr/bin/env bash

set -uo pipefail

project_root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
backend_pid=""
frontend_pid=""

cleanup() {
	trap - INT TERM EXIT

	for pid in "$backend_pid" "$frontend_pid"; do
		if [[ -n "$pid" ]] && kill -0 "$pid" 2>/dev/null; then
			kill "$pid" 2>/dev/null || true
		fi
	done

	wait "$backend_pid" "$frontend_pid" 2>/dev/null || true
}

trap cleanup INT TERM EXIT

(
	cd "$project_root/backend"
	set -a
	# shellcheck disable=SC1091
	source ./.env
	set +a
	exec go run ./cmd/api
) &
backend_pid=$!

(
	cd "$project_root/frontend"
	exec npm run dev
) &
frontend_pid=$!

echo "Backend i frontend su pokrenuti. Pritisni Ctrl+C za zaustavljanje."

wait -n "$backend_pid" "$frontend_pid"
status=$?

exit "$status"
