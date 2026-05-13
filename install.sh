#!/usr/bin/env bash
set -euo pipefail

REPO="dozheiny/barghman"
BINARY="barghman"
INSTALL_DIR="/usr/local/bin"

# ── colour helpers ────────────────────────────────────────────────────────────
RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'; NC='\033[0m'
info()  { echo -e "${GREEN}[barghman]${NC} $*"; }
warn()  { echo -e "${YELLOW}[barghman]${NC} $*"; }
error() { echo -e "${RED}[barghman]${NC} $*" >&2; exit 1; }

# ── detect OS ────────────────────────────────────────────────────────────────
case "$(uname -s)" in
  Linux*)   OS="linux" ;;
  Darwin*)  OS="darwin" ;;
  MINGW*|MSYS*|CYGWIN*) OS="windows" ;;
  *)        error "Unsupported OS: $(uname -s)" ;;
esac

# ── detect ARCH ───────────────────────────────────────────────────────────────
case "$(uname -m)" in
  x86_64|amd64)   ARCH="amd64" ;;
  arm64|aarch64)  ARCH="arm64" ;;
  *)              error "Unsupported architecture: $(uname -m)" ;;
esac

# ── resolve latest version ────────────────────────────────────────────────────
info "Fetching latest release..."
if command -v curl &>/dev/null; then
  FETCH="curl -fsSL"
elif command -v wget &>/dev/null; then
  FETCH="wget -qO-"
else
  error "Neither curl nor wget found. Please install one and retry."
fi

LATEST_URL="https://api.github.com/repos/${REPO}/releases/latest"
VERSION=$(${FETCH} "${LATEST_URL}" | grep '"tag_name"' | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/')
[ -z "${VERSION}" ] && error "Could not determine latest version."
info "Latest version: ${VERSION}"

# ── build download URL ────────────────────────────────────────────────────────
BASE_URL="https://github.com/${REPO}/releases/download/${VERSION}"

if [ "${OS}" = "windows" ]; then
  ARCHIVE="${BINARY}_${VERSION#v}_${OS}_${ARCH}.zip"
  EXT="zip"
else
  ARCHIVE="${BINARY}_${VERSION#v}_${OS}_${ARCH}.tar.gz"
  EXT="tar.gz"
fi

DOWNLOAD_URL="${BASE_URL}/${ARCHIVE}"
info "Downloading ${DOWNLOAD_URL}"

# ── download to tmp ───────────────────────────────────────────────────────────
TMP_DIR=$(mktemp -d)
trap 'rm -rf "${TMP_DIR}"' EXIT

${FETCH} "${DOWNLOAD_URL}" -o "${TMP_DIR}/${ARCHIVE}" 2>/dev/null \
  || ${FETCH} "${DOWNLOAD_URL}" > "${TMP_DIR}/${ARCHIVE}"

# ── extract ───────────────────────────────────────────────────────────────────
info "Extracting..."
if [ "${EXT}" = "zip" ]; then
  command -v unzip &>/dev/null || error "unzip is required on Windows/MSYS."
  unzip -q "${TMP_DIR}/${ARCHIVE}" -d "${TMP_DIR}"
else
  tar -xzf "${TMP_DIR}/${ARCHIVE}" -C "${TMP_DIR}"
fi

EXTRACTED_BINARY="${TMP_DIR}/${BINARY}"
[ "${OS}" = "windows" ] && EXTRACTED_BINARY="${TMP_DIR}/${BINARY}.exe"
[ -f "${EXTRACTED_BINARY}" ] || error "Binary not found after extraction."
chmod +x "${EXTRACTED_BINARY}"

# ── install ───────────────────────────────────────────────────────────────────
if [ "${OS}" = "windows" ]; then
  INSTALL_DIR="${USERPROFILE}/bin"
  mkdir -p "${INSTALL_DIR}"
  cp "${EXTRACTED_BINARY}" "${INSTALL_DIR}/${BINARY}.exe"
  info "Installed to ${INSTALL_DIR}/${BINARY}.exe"
  warn "Make sure ${INSTALL_DIR} is in your PATH."
else
  if [ -w "${INSTALL_DIR}" ]; then
    cp "${EXTRACTED_BINARY}" "${INSTALL_DIR}/${BINARY}"
  else
    info "Requesting sudo to install to ${INSTALL_DIR}..."
    sudo cp "${EXTRACTED_BINARY}" "${INSTALL_DIR}/${BINARY}"
  fi
  info "Installed to ${INSTALL_DIR}/${BINARY}"
fi

# ── verify ────────────────────────────────────────────────────────────────────
if command -v "${BINARY}" &>/dev/null; then
  info "Installation complete! Run: ${BINARY} -file <config.toml>"
else
  warn "Installed, but '${BINARY}' is not in PATH yet."
  warn "Add ${INSTALL_DIR} to your PATH and restart your shell."
fi