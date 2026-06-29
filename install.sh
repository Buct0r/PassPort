#!/usr/bin/env bash
#
# PassPort - Install Script
# https://github.com/Buct0r/PassPort
#
# Install the PassPort password manager on Linux and macOS.
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/Buct0r/PassPort/main/install.sh | bash
#   curl -fsSL https://raw.githubusercontent.com/Buct0r/PassPort/main/install.sh | bash -s -- --cli
#   ./install.sh --gui
#   ./install.sh --prefix ~/.local --version v0.3.0
#
# Options:
#   --cli              Install CLI only (non-interactive default)
#   --gui              Install GUI only (non-interactive)
#   --prefix <dir>     Install to <dir>/bin (default: /usr/local)
#   --version <tag>    Install specific version (default: latest)
#   --yes, -y          Non-interactive mode (answer yes to prompts)
#   --help, -h         Show this help

set -euo pipefail
IFS=$'\n\t'

REPO="Buct0r/PassPort"
PROJECT="PassPort"
BIN_GUI="passport"
BIN_CLI="passport-cli"
DEFAULT_PREFIX="/usr/local"

NC='\033[0m'
RED='\033[0;31m'
GREEN='\033[0;32m'
CYAN='\033[0;36m'
YELLOW='\033[1;33m'

info()  { printf "${CYAN}%s${NC}\n" "$*"; }
ok()    { printf "${GREEN}%s${NC}\n" "$*"; }
warn()  { printf "${YELLOW}%s${NC}\n" "$*"; }
error() { printf "${RED}Error: %s${NC}\n" "$*" >&2; exit 1; }

usage() {
    cat <<'EOF'
PassPort Install Script

Usage: install.sh [options]

Options:
  --cli              Install CLI only (non-interactive default)
  --gui              Install GUI only (non-interactive)
  --prefix <dir>     Install to <dir>/bin (default: /usr/local)
  --version <tag>    Install specific version (e.g. v0.3.0)
  --yes, -y          Non-interactive mode
  --help, -h         Show this help
EOF
    exit 0
}

has_cmd() { command -v "$1" >/dev/null 2>&1; }

cleanup() {
    [ -n "${TMPDIR:-}" ] && rm -rf "$TMPDIR"
}
trap cleanup EXIT

fetch_url() {
    if has_cmd curl; then
        curl -fsSL "$1"
    elif has_cmd wget; then
        wget -qO- "$1"
    else
        error "Neither curl nor wget found. Install one of them and retry."
    fi
}

download_file() {
    local url="$1" out="$2"
    if has_cmd curl; then
        curl -fsSL "$url" -o "$out"
    elif has_cmd wget; then
        wget -qO "$out" "$url"
    else
        error "Neither curl nor wget found."
    fi
}

parse_args() {
    INSTALL_CLI=false
    INSTALL_GUI=false
    PREFIX="$DEFAULT_PREFIX"
    VERSION=""
    NONINTERACTIVE=false

    while [ $# -gt 0 ]; do
        case "$1" in
            --cli)      INSTALL_CLI=true ;;
            --gui)      INSTALL_GUI=true ;;
            --prefix)   shift; PREFIX="$1" ;;
            --version)  shift; VERSION="$1" ;;
            --yes|-y)   NONINTERACTIVE=true ;;
            --help|-h)  usage ;;
            *)          error "Unknown option: $1 (use --help)" ;;
        esac
        shift
    done

    if $INSTALL_CLI && $INSTALL_GUI; then
        error "Cannot specify both --cli and --gui"
    fi
}

detect_platform() {
    OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
    ARCH="$(uname -m)"

    case "$ARCH" in
        x86_64|amd64) ARCH="amd64" ;;
        aarch64|arm64) ARCH="arm64" ;;
        *) error "Unsupported architecture: $ARCH (expected amd64 or arm64)" ;;
    esac

    case "$OS" in
        linux|darwin) ;;
        mingw*|msys*|cygwin*) error "This script does not support Windows. Use winget or the installer from the releases page." ;;
        *) error "Unsupported OS: $OS" ;;
    esac
}

detect_variants() {
    GUI_AVAIL=false
    CLI_AVAIL=true

    # Linux amd64 has GUI+CLI; Linux arm64 has CLI only
    if [ "$OS" = "linux" ] && [ "$ARCH" = "amd64" ]; then
        GUI_AVAIL=true
    fi
}

check_gui_deps() {
    if ! $INSTALL_GUI && [ "$VARIANT" != "gui" ] && [ "$VARIANT" != "both" ]; then
        return
    fi

    if [ "$OS" != "linux" ]; then
        return
    fi

    if has_cmd dpkg; then
        local missing=""
        dpkg -l libgl1-mesa-dev >/dev/null 2>&1 || missing="$missing libgl1-mesa-dev"
        dpkg -l xorg-dev >/dev/null 2>&1 || missing="$missing xorg-dev"
        if [ -n "$missing" ]; then
            warn "Missing GUI dependencies:$missing"
            warn "Install with: sudo apt-get install$missing"
            $NONINTERACTIVE && error "Missing GUI dependencies. Install them first or use --cli."
            echo ""
            if ! confirm "Continue anyway? (GUI may not work)"; then
                error "Aborted."
            fi
        fi
    elif has_cmd rpm; then
        local missing=""
        rpm -q mesa-libGL-devel >/dev/null 2>&1 || missing="$missing mesa-libGL-devel"
        rpm -q libXcursor-devel >/dev/null 2>&1 || missing="$missing libXcursor-devel"
        rpm -q libXrandr-devel >/dev/null 2>&1 || missing="$missing libXrandr-devel"
        rpm -q libXinerama-devel >/dev/null 2>&1 || missing="$missing libXinerama-devel"
        if [ -n "$missing" ]; then
            warn "Missing GUI dependencies:$missing"
            warn "Install with: sudo dnf install$missing"
            $NONINTERACTIVE && error "Missing GUI dependencies. Install them first or use --cli."
            echo ""
            if ! confirm "Continue anyway? (GUI may not work)"; then
                error "Aborted."
            fi
        fi
    else
        warn "Cannot verify GUI dependencies (unsupported package manager)."
        warn "Ensure OpenGL development libraries are installed if the GUI fails to start."
    fi
}

confirm() {
    local prompt="${1:-Continue?}"
    local reply
    read -r -p "$prompt [y/N] " reply
    case "$reply" in
        [yY]|[yY][eE][sS]) return 0 ;;
        *) return 1 ;;
    esac
}

choose_variant() {
    if $INSTALL_CLI; then
        VARIANT="cli"
        return
    fi
    if $INSTALL_GUI; then
        $GUI_AVAIL || error "GUI is not available for $OS/$ARCH"
        VARIANT="gui"
        return
    fi

    # Non-interactive default: CLI
    if [ ! -t 0 ] || $NONINTERACTIVE; then
        info "Non-interactive mode: installing CLI only."
        info "Use --gui or --cli to select explicitly."
        VARIANT="cli"
        return
    fi

    echo ""
    echo "Which version would you like to install?"
    if $GUI_AVAIL; then
        echo "  1) GUI + CLI (recommended)"
        echo "  2) CLI only"
        read -r -p "Choice [1]: " choice
        case "${choice:-1}" in
            1) VARIANT="both" ;;
            2) VARIANT="cli" ;;
            *) error "Invalid choice: $choice" ;;
        esac
    else
        info "CLI-only install (GUI not available for $OS/$ARCH)"
        VARIANT="cli"
    fi
}

fetch_version() {
    if [ -n "$VERSION" ]; then
        ok "Using specified version: $VERSION"
        return
    fi

    info "Fetching latest release..."
    local json
    json="$(fetch_url "https://api.github.com/repos/$REPO/releases/latest")" || \
        error "Failed to fetch latest release (rate limited?). Use --version to specify a tag."

    VERSION="$(echo "$json" | grep -o '"tag_name": *"[^"]*"' | cut -d'"' -f4)"
    [ -z "$VERSION" ] && error "Could not determine latest version. Use --version to specify a tag."

    ok "Latest release: $VERSION"
}

download_and_extract() {
    local variant="$1"
    local ver="${VERSION#v}"
    local archive="${PROJECT}-${variant}-${ver}-${OS}-${ARCH}.tar.gz"
    local url="https://github.com/$REPO/releases/download/$VERSION/$archive"

    info "Downloading $archive ..."
    download_file "$url" "$TMPDIR/$archive"

    info "Extracting $archive ..."
    tar -xzf "$TMPDIR/$archive" -C "$TMPDIR"
}

do_install() {
    local bin_dir="${PREFIX}/bin"

    if [ ! -d "$bin_dir" ]; then
        if [ "$PREFIX" = "$DEFAULT_PREFIX" ]; then
            error "$bin_dir does not exist. Use --prefix to install elsewhere (e.g. --prefix ~/.local)."
        fi
        mkdir -p "$bin_dir"
    fi

    local use_sudo=false
    if [ "$PREFIX" = "$DEFAULT_PREFIX" ] && [ "$(id -u)" -ne 0 ]; then
        has_cmd sudo || error "sudo is required to install to $PREFIX. Use --prefix to install elsewhere."
        use_sudo=true
    fi

    for binary in "$@"; do
        if [ ! -f "$TMPDIR/$binary" ]; then
            error "Binary $binary not found in archive (expected: $TMPDIR/$binary)"
        fi

        if $use_sudo; then
            sudo install -m 0755 "$TMPDIR/$binary" "$bin_dir/"
        else
            install -m 0755 "$TMPDIR/$binary" "$bin_dir/"
        fi
        ok "Installed $binary to $bin_dir/"
    done
}

main() {
    parse_args "$@"
    detect_platform
    detect_variants
    choose_variant
    check_gui_deps
    fetch_version

    TMPDIR="$(mktemp -d)"

    case "$VARIANT" in
        both)
            download_and_extract "gui"
            download_and_extract "cli"
            do_install "$BIN_GUI" "$BIN_CLI"
            ;;
        gui)
            download_and_extract "gui"
            do_install "$BIN_GUI"
            info "TIP: The GUI can launch the CLI with 'passport --cli' if passport-cli is also in your PATH."
            ;;
        cli)
            download_and_extract "cli"
            do_install "$BIN_CLI"
            ;;
    esac

    echo ""
    ok "PassPort $VERSION installed successfully!"
    info "  GUI:  run 'passport'"
    info "  CLI:  run 'passport-cli' or 'passport --cli'"
    echo ""
}

main "$@"
