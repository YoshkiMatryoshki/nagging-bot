#!/usr/bin/env bash
set -euo pipefail

ROOT_PATH="${1:-}"
SERVICE_NAME="${2:-}"
BINARY_NAME="${3:-naggingbot}"
BACKUP_DB="${4:-false}"

if [[ -z "${ROOT_PATH}" || -z "${SERVICE_NAME}" ]]; then
  echo "usage: deploy_remote.sh <root_path> <service_name> [binary_name] [backup_db:true|false]" >&2
  exit 2
fi

case "${BACKUP_DB}" in
  true|false) ;;
  *)
    echo "invalid backup_db value: ${BACKUP_DB}. expected true|false" >&2
    exit 2
    ;;
esac

RELEASE_ID="$(date -u +%Y%m%dT%H%M%SZ)"
RELEASE_DIR="${ROOT_PATH}/releases/${RELEASE_ID}"
UPLOAD_BIN="${ROOT_PATH}/upload/${BINARY_NAME}"
CURRENT_DIR="${ROOT_PATH}/current"
CURRENT_BIN="${CURRENT_DIR}/${BINARY_NAME}"
TMP_BIN="${CURRENT_DIR}/.${BINARY_NAME}.tmp.$$"
SHARED_DIR="${ROOT_PATH}/shared"
SHARED_ENV="${SHARED_DIR}/.env"
SHARED_SERVICE="${SHARED_DIR}/${SERVICE_NAME}.service"
SYSTEMD_SERVICE="/etc/systemd/system/${SERVICE_NAME}.service"
DB_PATH="${SHARED_DIR}/naggingbot.db"

if [[ ! -f "${UPLOAD_BIN}" ]]; then
  echo "missing upload binary: ${UPLOAD_BIN}" >&2
  exit 1
fi
if [[ ! -f "${SHARED_ENV}" ]]; then
  echo "missing env file: ${SHARED_ENV}" >&2
  exit 1
fi
if [[ ! -f "${SHARED_SERVICE}" ]]; then
  echo "missing service file: ${SHARED_SERVICE}" >&2
  exit 1
fi

mkdir -p "${ROOT_PATH}/releases" "${CURRENT_DIR}" "${SHARED_DIR}" "${ROOT_PATH}/upload" "${RELEASE_DIR}"

# Stop active unit if it already exists.
if sudo systemctl is-active --quiet "${SERVICE_NAME}.service"; then
  sudo systemctl stop "${SERVICE_NAME}.service"
fi

# Backup current binary and optional DB snapshot for rollback.
if [[ -f "${CURRENT_BIN}" ]]; then
  cp -a "${CURRENT_BIN}" "${RELEASE_DIR}/${BINARY_NAME}.prev"
fi
if [[ "${BACKUP_DB}" == "true" && -f "${DB_PATH}" ]]; then
  cp -a "${DB_PATH}" "${RELEASE_DIR}/naggingbot.db.bak"
fi

ln -sfn "${SHARED_ENV}" "${CURRENT_DIR}/.env"
install -m 0755 "${UPLOAD_BIN}" "${TMP_BIN}"
mv -f "${TMP_BIN}" "${CURRENT_BIN}"

if [[ ! -x "${CURRENT_BIN}" || ! -s "${CURRENT_BIN}" ]]; then
  echo "current binary is missing or not executable: ${CURRENT_BIN}" >&2
  exit 1
fi

if [[ ! -f "${SYSTEMD_SERVICE}" ]] || ! cmp -s "${SHARED_SERVICE}" "${SYSTEMD_SERVICE}"; then
  sudo install -m 0644 "${SHARED_SERVICE}" "${SYSTEMD_SERVICE}"
  sudo systemctl daemon-reload
fi

sudo systemctl start "${SERVICE_NAME}.service"

if ! sudo systemctl is-active --quiet "${SERVICE_NAME}.service"; then
  sudo systemctl status "${SERVICE_NAME}.service" --no-pager || true
  sudo journalctl -u "${SERVICE_NAME}.service" -n 200 --no-pager || true
  exit 1
fi

echo "deploy success: service=${SERVICE_NAME} release=${RELEASE_ID}"
