#!/bin/bash

# Script pour télécharger et installer tfplugindocs
# Usage: ./bin/down.sh

# $1 text black + background bright grey then $2 ... in default colors
log() { printf "\033[48;05;255m\033[38;05;0m ${1} \033[0m "; echo ${@:2}; }
# $1 text black + background bright green then $2 ... in default colors
info() { printf "\033[48;05;118m\033[38;05;0m ${1} \033[0m "; echo ${@:2}; }
# $1 text black + background orange then $2 ... in default colors
warn() { printf "\033[48;05;214m\033[38;05;0m ${1} \033[0m "; echo ${@:2}; }
# $1 text white + background red then $2 ... in default colors
error() { printf "\033[48;05;196m\033[38;05;255m ${1} \033[0m "; echo ${@:2}; }

# Variables globales
VERBOSE=
BIN_DIR="$(cd "$(dirname "$0")" && pwd)"

# Parse les arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -v|--verbose) VERBOSE=1; shift ;;
        *) error ABORT "Unknown option: $1" && exit 1 ;;
    esac
done

# Déterminer l'OS et l'architecture
detect_platform() {
    case "$(uname -s)" in
        Darwin*) OS="darwin" ;;
        Linux*) OS="linux" ;;
        MINGW*|MSYS*|CYGWIN*) OS="windows" ;;
        *) error ABORT "Unsupported OS: $(uname -s)" && exit 1 ;;
    esac
    [[ -n "$VERBOSE" ]] && log OS "$OS"

    case "$(uname -m)" in
        x86_64|amd64) ARCH="amd64" ;;
        arm64|aarch64) ARCH="arm64" ;;
        armv7l) ARCH="arm" ;;
        *) error ABORT "Unsupported Arch: $(uname -m)" && exit 1 ;;
    esac
    [[ -n "$VERBOSE" ]] && log ARCH "$ARCH"
}

# Récupère le nom de la dernière release sur GitHub
get_latest_version() {
    LATEST_VERSION=$(curl -s https://api.github.com/repos/hashicorp/terraform-plugin-docs/releases/latest | jq -r '.tag_name // ""' | sed 's/^v//')
    [[ -n "$VERBOSE" ]] && log LATEST_VERSION $LATEST_VERSION

    if [ -z "$LATEST_VERSION" ]; then
        error ABORT "Failed to get latest version from GitHub"
        exit 1
    fi
}

# Vérifie la version locale de tfplugindocs
check_existing_version() {
    if [[ -f "$BIN_DIR/tfplugindocs" ]]; then
        CURRENT_VERSION=$("$BIN_DIR/tfplugindocs" --version 2>&1 | grep -o '[0-9]\+\.[0-9]\+\.[0-9]\+' | sed 's/^v//')
        [[ -n "$VERBOSE" ]] && log CURRENT_VERSION $CURRENT_VERSION

        if [ "$CURRENT_VERSION" = "$LATEST_VERSION" ]; then
            info INFO "tfplugindocs already up to date ($LATEST_VERSION)"
            exit 0
        else
            [[ -n "$VERBOSE" ]] && log INFO "Updating tfplugindocs from $CURRENT_VERSION to $LATEST_VERSION"
        fi
    else
        [[ -n "$VERBOSE" ]] && log INFO "tfplugindocs not found, will install $LATEST_VERSION"
    fi
}

# Télécharger et installer tfplugindocs
download_and_install() {
    ASSET_NAME="tfplugindocs_${LATEST_VERSION}_${OS}_${ARCH}.zip"
    DOWNLOAD_URL=$(curl -s https://api.github.com/repos/hashicorp/terraform-plugin-docs/releases/latest | jq -r --arg name "$ASSET_NAME" '.assets[] | select(.name == $name) | .browser_download_url')
    
    if [ -z "$DOWNLOAD_URL" ] || [ "$DOWNLOAD_URL" = "null" ]; then
        error ABORT "$ASSET_NAME not found"
        exit 1
    fi
    
    # Download
    TEMP_DIR=$(mktemp -d)
    cd "$TEMP_DIR"

    curl --silent --location --fail --output tfplugindocs.zip "$DOWNLOAD_URL"

    # Extraire
    unzip -q tfplugindocs.zip
    
    # Vérifie
    if [[ ! -f "tfplugindocs" ]]; then
        error ABORT tfplugindocs not found
        rm -rf "$TEMP_DIR"
        exit 1
    fi
    
    chmod +x tfplugindocs
    
    # Installe dans le même dossier que le script
    [[ -n "$VERBOSE" ]] && log INFO "Moving tfplugindocs to $BIN_DIR"
    mv "$TEMP_DIR/tfplugindocs" "$BIN_DIR"
    
    rm -rf "$TEMP_DIR"
    
    # Vérifier que le fichier a bien été installé
    if [[ -f "$BIN_DIR/tfplugindocs" ]]; then
        CURRENT_VERSION=$("$BIN_DIR/tfplugindocs" --version 2>&1 | grep -o '[0-9]\+\.[0-9]\+\.[0-9]\+' | sed 's/^v//')
        info INFO "tfplugindocs $CURRENT_VERSION installed successfully"
    else
        error ABORT tfplugindocs installation failed - file not found
        exit 1
    fi
}

detect_platform
get_latest_version
check_existing_version
download_and_install
