#!/bin/bash

set -e

# --- Colors for output ---
green='\033[0;32m'
red='\033[0;31m'
yellow='\033[0;33m'
cyan='\033[0;36m'
nc='\033[0m' # No Color

# --- Configuration ---
REPO_OWNER="EliasObeid9-02"
REPO_NAME="CommitGen"
RELEASE_VERSION="v0.1.1"
BINARY_NAME="commitgen"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}" # Default install directory, can be overridden by user
LOCAL_INSTALL_DIR="${HOME}/.local/bin"
TEMP_DIR=$(mktemp -d)

# --- Helper Functions ---
log_info() {
  echo -e "${cyan}[INFO]${nc} $1"
}

log_success() {
  echo -e "${green}[SUCCESS]${nc} $1"
}

log_warning() {
  echo -e "${yellow}[WARNING]${nc} $1"
}

log_error() {
  echo -e "${red}[ERROR]${nc} $1" >&2
  exit 1
}

# --- Main Installation Logic ---
log_info "Starting CommitGen ${RELEASE_VERSION} installation..."

# 1. Detect OS and Architecture
OS=$(uname -s)
ARCH=$(uname -m)

case "${ARCH}" in
  x86_64|amd64)
    ARCH="amd64"
    ;;
  *)
    log_error "Unsupported architecture: ${ARCH}. This app only supports x86_64/amd64."
    ;;
esac

case "${OS}" in
  Linux)
    OS="linux"
    ;;
  *)
    log_error "Unsupported operating system: ${OS}. This app currently only supports Linux."
    ;;
esac

log_info "Detected OS: ${OS}, Architecture: ${ARCH}"

# 2. Construct download URL
TARBALL_FILENAME="${BINARY_NAME}-${OS}-${ARCH}.tar.gz"
DOWNLOAD_URL="https://github.com/${REPO_OWNER}/${REPO_NAME}/releases/download/${RELEASE_VERSION}/${TARBALL_FILENAME}"

log_info "Downloading ${TARBALL_FILENAME} from ${DOWNLOAD_URL} to ${TEMP_DIR}..."

DOWNLOAD_COMMAND=""
if command -v curl >/dev/null 2>&1; then
  DOWNLOAD_COMMAND="curl -L -o '${TEMP_DIR}/${TARBALL_FILENAME}' '${DOWNLOAD_URL}'"
elif command -v wget >/dev/null 2>&1; then
  DOWNLOAD_COMMAND="wget -q -O '${TEMP_DIR}/${TARBALL_FILENAME}' '${DOWNLOAD_URL}'"
else
  log_error "Neither curl nor wget found. Please install one to proceed."
fi

eval "${DOWNLOAD_COMMAND}" || log_error "Failed to download ${TARBALL_FILENAME}."

# 3. Extract and install
log_info "Extracting ${TARBALL_FILENAME}..."
if ! command -v tar >/dev/null 2>&1; then
  log_error "Dependency 'tar' not found. Please install it to proceed."
fi
tar -xzf "${TEMP_DIR}/${TARBALL_FILENAME}" -C "${TEMP_DIR}" || log_error "Failed to extract ${TARBALL_FILENAME}."

# The extracted binary will be named commitgen-linux-amd64
EXTRACTED_BINARY_NAME="${BINARY_NAME}-${OS}-${ARCH}"

log_info "Installing ${BINARY_NAME}..."

# Try to install to /usr/local/bin first
log_info "Press Ctrl+C to install locally to '${LOCAL_INSTALL_DIR}'"
if sudo mv "${TEMP_DIR}/${EXTRACTED_BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}" 2>/dev/null; then
  log_success "${BINARY_NAME} installed to ${INSTALL_DIR}"
else
  log_warning "Unable to move binary to ${INSTALL_DIR}. Attempting local installation to ~/.local/bin."
  mkdir -p "${LOCAL_INSTALL_DIR}" || log_error "Failed to create local installation directory: ${LOCAL_INSTALL_DIR}."
  if mv "${TEMP_DIR}/${EXTRACTED_BINARY_NAME}" "${LOCAL_INSTALL_DIR}/${BINARY_NAME}"; then
    log_success "${BINARY_NAME} installed to ${LOCAL_INSTALL_DIR}"
    # Add to PATH if not already there
    if ! [[ ":$PATH:" == "*:${LOCAL_INSTALL_DIR}:*" ]]; then
      log_info "Adding ${LOCAL_INSTALL_DIR} to PATH..."
      case $SHELL in
        */bash)
          echo 'export PATH="${HOME}/.local/bin":${PATH}' >> ~/.bashrc
          ;;
        */zsh)
          echo 'export PATH="${HOME}/.local/bin":${PATH}' >> ~/.zshrc
          ;;
        */fish)
          echo 'fish_add_path "${HOME}/.local/bin"' >> ~/.config/fish/config.fish
          ;;
        *)
          log_warning "Unsupported shell: ${SHELL}. Please manually add ${LOCAL_INSTALL_DIR} to your PATH."
          ;;
      esac
      log_warning "Please source your shell config file or restart your terminal for changes to take effect."
    fi
  else
    log_error "Failed to install ${BINARY_NAME} to both ${INSTALL_DIR} and ${LOCAL_INSTALL_DIR}."
  fi
fi

log_info "Setting executable permissions..."
# Check which path was used for installation and apply chmod there
if [ -f "${INSTALL_DIR}/${BINARY_NAME}" ]; then
  chmod +x "${INSTALL_DIR}/${BINARY_NAME}" || log_error "Failed to set executable permissions for ${BINARY_NAME} in ${INSTALL_DIR}."
elif [ -f "${LOCAL_INSTALL_DIR}/${BINARY_NAME}" ]; then
  chmod +x "${LOCAL_INSTALL_DIR}/${BINARY_NAME}" || log_error "Failed to set executable permissions for ${BINARY_NAME} in ${LOCAL_INSTALL_DIR}."
else
  log_error "${BINARY_NAME} not found after installation. Permissions could not be set."
fi

# 4. Cleanup
log_info "Cleaning up temporary files..."
rm -rf "${TEMP_DIR}"

log_success "CommitGen ${RELEASE_VERSION} installed successfully!"
log_info "You can now run '${BINARY_NAME}' from your terminal."
log_warning "Note: This is a pre-release version (${RELEASE_VERSION})."
