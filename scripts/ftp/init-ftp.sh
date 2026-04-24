#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
VENDOR_DIR="${SCRIPT_DIR}/.vendor"
WORKSPACE_DIR="${REPO_ROOT}/integration/ftp-workspace"
PID_FILE="${SCRIPT_DIR}/ftp-server.pid"
LOG_FILE="${SCRIPT_DIR}/ftp-server.log"
HOST="127.0.0.1"
PORT="2121"

mkdir -p "${VENDOR_DIR}" "${WORKSPACE_DIR}"

if ! python3 -m pip --version >/dev/null 2>&1; then
	python3 -m ensurepip --upgrade --user >/dev/null 2>&1 || {
		echo "python3 pip is required to install pyftpdlib; failed to bootstrap with ensurepip" >&2
		exit 1
	}
fi

python3 -m pip install --disable-pip-version-check --quiet --target "${VENDOR_DIR}" pyftpdlib

if [ -f "${PID_FILE}" ]; then
	OLD_PID="$(tr -d '[:space:]' < "${PID_FILE}")"
	if [ -n "${OLD_PID}" ] && kill -0 "${OLD_PID}" 2>/dev/null; then
		kill "${OLD_PID}" 2>/dev/null || true
		for _ in $(seq 1 10); do
			if ! kill -0 "${OLD_PID}" 2>/dev/null; then
				break
			fi
			sleep 1
		done
	fi
	rm -f "${PID_FILE}"
fi

rm -rf "${WORKSPACE_DIR}"
mkdir -p "${WORKSPACE_DIR}"
mkdir -p "${WORKSPACE_DIR}/ftp-push" "${WORKSPACE_DIR}/ftp-pull"

PYTHONPATH="${VENDOR_DIR}" nohup python3 "${SCRIPT_DIR}/server.py" \
	--host "${HOST}" \
	--port "${PORT}" \
	--workspace "${WORKSPACE_DIR}" \
	--user "ftp_user" \
	--password "ftp_pwd" \
	--passive-ports "30000-30009" \
	>"${LOG_FILE}" 2>&1 &

SERVER_PID=$!
printf '%s\n' "${SERVER_PID}" > "${PID_FILE}"

for _ in $(seq 1 30); do
	if PYTHONPATH="${VENDOR_DIR}" python3 - <<'PY'
import socket

s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
s.settimeout(0.5)
try:
	result = s.connect_ex(("127.0.0.1", 2121))
finally:
	s.close()
raise SystemExit(0 if result == 0 else 1)
PY
	then
		echo "FTP test server started at ${HOST}:${PORT}"
		echo "Workspace: ${WORKSPACE_DIR}"
		echo "PID file: ${PID_FILE}"
		echo "Log file: ${LOG_FILE}"
		exit 0
	fi
	if ! kill -0 "${SERVER_PID}" 2>/dev/null; then
		echo "FTP test server exited unexpectedly. See ${LOG_FILE}" >&2
		exit 1
	fi
	sleep 1
done

echo "Timed out waiting for FTP test server on ${HOST}:${PORT}. See ${LOG_FILE}" >&2
	exit 1
