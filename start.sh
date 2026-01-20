#!/bin/bash

set -e

# Colors for fallback menu
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m'

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

OLLAMA_KEY_FILE="$SCRIPT_DIR/.ollama_api_key"
SYNTHESIA_KEY_FILE="$SCRIPT_DIR/.synthesia_api_key"
UNSPLASH_KEY_FILE="$SCRIPT_DIR/.unsplash_api_key"

# Get or prompt for Ollama Cloud API key
get_ollama_api_key() {
    local key=""

    # Try to load existing key
    if [[ -f "$OLLAMA_KEY_FILE" ]]; then
        key=$(cat "$OLLAMA_KEY_FILE" 2>/dev/null)
        if [[ -n "$key" ]]; then
            echo "$key"
            return 0
        fi
    fi

    return 1
}

# Prompt for API key (TUI version) - sets OLLAMA_API_KEY variable
prompt_api_key_tui() {
    local cmd=$1
    OLLAMA_API_KEY=""

    # Check for existing key
    if OLLAMA_API_KEY=$(get_ollama_api_key); then
        if $cmd --title "Ollama Cloud" --yesno "Found saved API key.\n\nUse existing key?" 10 50; then
            return 0
        fi
        OLLAMA_API_KEY=""
    fi

    # Prompt for new key
    OLLAMA_API_KEY=$($cmd --title "Ollama Cloud API Key" \
        --inputbox "Enter your Ollama API key:\n\n(Get one at https://ollama.com/settings/keys)" 12 60 \
        3>&1 1>&2 2>&3) || return 1

    if [[ -z "$OLLAMA_API_KEY" ]]; then
        $cmd --title "Error" --msgbox "API key is required for Ollama Cloud." 8 50
        return 1
    fi

    # Save the key
    echo "$OLLAMA_API_KEY" > "$OLLAMA_KEY_FILE"
    chmod 600 "$OLLAMA_KEY_FILE"
}

# Prompt for API key (simple version)
prompt_api_key_simple() {
    local key=""

    # Check for existing key
    if key=$(get_ollama_api_key); then
        echo -e "${GREEN}Found saved API key.${NC}"
        read -p "Use existing key? [Y/n]: " use_existing
        if [[ ! "$use_existing" =~ ^[Nn]$ ]]; then
            echo "$key"
            return 0
        fi
    fi

    # Prompt for new key
    echo ""
    echo -e "${CYAN}Get your API key at: https://ollama.com/settings/keys${NC}"
    echo ""
    read -p "Enter Ollama API key: " key

    if [[ -z "$key" ]]; then
        echo -e "${RED}API key is required for Ollama Cloud.${NC}"
        return 1
    fi

    # Save the key
    echo "$key" > "$OLLAMA_KEY_FILE"
    chmod 600 "$OLLAMA_KEY_FILE"
    echo "$key"
}

# Get or prompt for Synthesia API key
get_synthesia_api_key() {
    local key=""

    # Try to load existing key
    if [[ -f "$SYNTHESIA_KEY_FILE" ]]; then
        key=$(cat "$SYNTHESIA_KEY_FILE" 2>/dev/null)
        if [[ -n "$key" ]]; then
            echo "$key"
            return 0
        fi
    fi

    return 1
}

# Prompt for Synthesia API key (TUI version) - sets SYNTHESIA_API_KEY variable
prompt_synthesia_key_tui() {
    local cmd=$1
    SYNTHESIA_API_KEY=""

    # Check for existing key
    if SYNTHESIA_API_KEY=$(get_synthesia_api_key); then
        if $cmd --title "Synthesia" --yesno "Found saved Synthesia API key.\n\nUse existing key?" 10 50; then
            return 0
        fi
        SYNTHESIA_API_KEY=""
    fi

    # Ask if they want to configure Synthesia
    if ! $cmd --title "Synthesia Video Generation" --yesno "Configure Synthesia for AI video generation?\n\n(Optional - courses will work without it)" 10 55; then
        return 0
    fi

    # Prompt for new key
    SYNTHESIA_API_KEY=$($cmd --title "Synthesia API Key" \
        --inputbox "Enter your Synthesia API key:\n\n(Get one at synthesia.io)" 12 60 \
        3>&1 1>&2 2>&3) || return 0

    if [[ -n "$SYNTHESIA_API_KEY" ]]; then
        # Save the key
        echo "$SYNTHESIA_API_KEY" > "$SYNTHESIA_KEY_FILE"
        chmod 600 "$SYNTHESIA_KEY_FILE"
    fi
}

# Prompt for Synthesia API key (simple version)
prompt_synthesia_key_simple() {
    local key=""

    # Check for existing key
    if key=$(get_synthesia_api_key); then
        echo -e "${GREEN}Found saved Synthesia API key.${NC}"
        read -p "Use existing key? [Y/n]: " use_existing
        if [[ ! "$use_existing" =~ ^[Nn]$ ]]; then
            echo "$key"
            return 0
        fi
    fi

    # Ask if they want to configure Synthesia
    echo ""
    read -p "Configure Synthesia for AI video generation? (optional) [y/N]: " configure
    if [[ ! "$configure" =~ ^[Yy]$ ]]; then
        echo ""
        return 0
    fi

    # Prompt for new key
    echo ""
    echo -e "${CYAN}Get your API key at: https://synthesia.io${NC}"
    echo ""
    read -p "Enter Synthesia API key (or press Enter to skip): " key

    if [[ -n "$key" ]]; then
        # Save the key
        echo "$key" > "$SYNTHESIA_KEY_FILE"
        chmod 600 "$SYNTHESIA_KEY_FILE"
    fi
    echo "$key"
}

# Get or prompt for Unsplash API key
get_unsplash_api_key() {
    local key=""

    # Try to load existing key
    if [[ -f "$UNSPLASH_KEY_FILE" ]]; then
        key=$(cat "$UNSPLASH_KEY_FILE" 2>/dev/null)
        if [[ -n "$key" ]]; then
            echo "$key"
            return 0
        fi
    fi

    return 1
}

# Prompt for Unsplash API key (TUI version) - sets UNSPLASH_API_KEY variable
prompt_unsplash_key_tui() {
    local cmd=$1
    UNSPLASH_API_KEY=""

    # Check for existing key
    if UNSPLASH_API_KEY=$(get_unsplash_api_key); then
        if $cmd --title "Unsplash" --yesno "Found saved Unsplash API key.\n\nUse existing key?" 10 50; then
            return 0
        fi
        UNSPLASH_API_KEY=""
    fi

    # Ask if they want to configure Unsplash
    if ! $cmd --title "Unsplash Stock Images" --yesno "Configure Unsplash for stock images in presentations?\n\n(Optional - presentations will work without it)" 10 55; then
        return 0
    fi

    # Prompt for new key
    UNSPLASH_API_KEY=$($cmd --title "Unsplash API Key" \
        --inputbox "Enter your Unsplash Access Key:\n\n(Get one at unsplash.com/developers)" 12 60 \
        3>&1 1>&2 2>&3) || return 0

    if [[ -n "$UNSPLASH_API_KEY" ]]; then
        # Save the key
        echo "$UNSPLASH_API_KEY" > "$UNSPLASH_KEY_FILE"
        chmod 600 "$UNSPLASH_KEY_FILE"
    fi
}

# Prompt for Unsplash API key (simple version)
prompt_unsplash_key_simple() {
    local key=""

    # Check for existing key
    if key=$(get_unsplash_api_key); then
        echo -e "${GREEN}Found saved Unsplash API key.${NC}"
        read -p "Use existing key? [Y/n]: " use_existing
        if [[ ! "$use_existing" =~ ^[Nn]$ ]]; then
            echo "$key"
            return 0
        fi
    fi

    # Ask if they want to configure Unsplash
    echo ""
    read -p "Configure Unsplash for stock images in presentations? (optional) [y/N]: " configure
    if [[ ! "$configure" =~ ^[Yy]$ ]]; then
        echo ""
        return 0
    fi

    # Prompt for new key
    echo ""
    echo -e "${CYAN}Get your Access Key at: https://unsplash.com/developers${NC}"
    echo ""
    read -p "Enter Unsplash Access Key (or press Enter to skip): " key

    if [[ -n "$key" ]]; then
        # Save the key
        echo "$key" > "$UNSPLASH_KEY_FILE"
        chmod 600 "$UNSPLASH_KEY_FILE"
    fi
    echo "$key"
}

# Detect if native Ollama is running
check_native_ollama() {
    if curl -s --connect-timeout 2 http://localhost:11434/api/tags >/dev/null 2>&1; then
        return 0
    fi
    return 1
}

# Start services based on selection
start_services() {
    local ollama_mode=$1
    local ollama_key=$2
    local synthesia_key=$3
    local unsplash_key=$4

    echo -e "${CYAN}Starting SecuSense...${NC}"

    # Common environment variables
    local common_env="SYNTHESIA_APIKEY=\"$synthesia_key\" UNSPLASH_ACCESSKEY=\"$unsplash_key\""

    case $ollama_mode in
        docker)
            echo -e "${BLUE}Starting with Docker Ollama${NC}"
            OLLAMA_URL="http://ollama:11434" SYNTHESIA_APIKEY="$synthesia_key" UNSPLASH_ACCESSKEY="$unsplash_key" docker compose --profile with-ollama up -d --build --force-recreate
            ;;
        cloud)
            echo -e "${BLUE}Using Ollama Cloud${NC}"
            OLLAMA_URL="https://ollama.com" OLLAMA_CLOUDMODE=true OLLAMA_APIKEY="$ollama_key" SYNTHESIA_APIKEY="$synthesia_key" UNSPLASH_ACCESSKEY="$unsplash_key" docker compose up -d --build --force-recreate
            ;;
        *)
            echo -e "${BLUE}Using native Ollama (host.docker.internal:11434)${NC}"
            SYNTHESIA_APIKEY="$synthesia_key" UNSPLASH_ACCESSKEY="$unsplash_key" docker compose up -d --build --force-recreate
            ;;
    esac

    echo ""
    echo -e "${GREEN}${BOLD}SecuSense is running!${NC}"
    echo -e "  Frontend: ${CYAN}http://localhost${NC}"
    echo -e "  API:      ${CYAN}http://localhost:8080${NC}"
    case $ollama_mode in
        docker)
            echo -e "  Ollama:   ${CYAN}http://localhost:11434${NC} (Docker)"
            ;;
        cloud)
            echo -e "  Ollama:   ${CYAN}Ollama Cloud${NC} (no local GPU required)"
            ;;
        *)
            echo -e "  Ollama:   ${CYAN}http://localhost:11434${NC} (Native)"
            ;;
    esac
}

# TUI menu using whiptail or dialog
show_tui_menu() {
    local cmd=""
    local native_status=""
    local default_item="native"

    if check_native_ollama; then
        native_status=" (detected running)"
    else
        native_status=" (not detected)"
        default_item="docker"
    fi

    # Try whiptail first, then dialog
    if command -v whiptail &>/dev/null; then
        cmd="whiptail"
    elif command -v dialog &>/dev/null; then
        cmd="dialog"
    else
        return 1
    fi

    local choice
    choice=$($cmd --title "SecuSense Launcher" \
        --menu "Select Ollama configuration:\n\nNative Ollama$native_status" 18 60 4 \
        "native" "Use native Ollama (recommended if installed)" \
        "docker" "Start Ollama in Docker" \
        "cloud" "Use Ollama Cloud (no GPU required)" \
        "stop" "Stop all services" \
        3>&1 1>&2 2>&3) || exit 0

    # Handle stop separately (no API key prompts needed)
    if [[ "$choice" == "stop" ]]; then
        echo -e "${CYAN}Stopping SecuSense...${NC}"
        docker compose --profile with-ollama down
        echo -e "${GREEN}Stopped.${NC}"
        return
    fi

    # Prompt for Synthesia key (used by all modes) - sets SYNTHESIA_API_KEY
    prompt_synthesia_key_tui "$cmd"

    # Prompt for Unsplash key (used by all modes) - sets UNSPLASH_API_KEY
    prompt_unsplash_key_tui "$cmd"

    case $choice in
        native)
            if ! check_native_ollama; then
                $cmd --title "Warning" --yesno "Native Ollama not detected.\n\nMake sure Ollama is running:\n  ollama serve\n\nContinue anyway?" 12 50 || exit 0
            fi
            start_services "native" "" "$SYNTHESIA_API_KEY" "$UNSPLASH_API_KEY"
            ;;
        docker)
            start_services "docker" "" "$SYNTHESIA_API_KEY" "$UNSPLASH_API_KEY"
            ;;
        cloud)
            # Prompt for Ollama key - sets OLLAMA_API_KEY
            prompt_api_key_tui "$cmd" || exit 0
            start_services "cloud" "$OLLAMA_API_KEY" "$SYNTHESIA_API_KEY" "$UNSPLASH_API_KEY"
            ;;
    esac
}

# Fallback simple menu
show_simple_menu() {
    local native_status=""

    echo -e "${BOLD}${CYAN}"
    echo "╔═══════════════════════════════════════╗"
    echo "║       SecuSense Launcher              ║"
    echo "╚═══════════════════════════════════════╝"
    echo -e "${NC}"

    if check_native_ollama; then
        native_status="${GREEN}(running)${NC}"
    else
        native_status="${RED}(not detected)${NC}"
    fi

    echo -e "Native Ollama: $native_status"
    echo ""
    echo -e "${BOLD}Select an option:${NC}"
    echo ""
    echo "  1) Use native Ollama (recommended if installed)"
    echo "  2) Start Ollama in Docker"
    echo "  3) Use Ollama Cloud (no GPU required)"
    echo "  4) Stop all services"
    echo "  5) Exit"
    echo ""

    read -p "Enter choice [1-5]: " choice

    # Skip API prompts for stop/exit
    local synthesia_key=""
    local unsplash_key=""
    if [[ "$choice" =~ ^[1-3]$ ]]; then
        synthesia_key=$(prompt_synthesia_key_simple)
        unsplash_key=$(prompt_unsplash_key_simple)
    fi

    case $choice in
        1)
            if ! check_native_ollama; then
                echo -e "${RED}Warning: Native Ollama not detected.${NC}"
                read -p "Continue anyway? [y/N]: " confirm
                [[ "$confirm" =~ ^[Yy]$ ]] || exit 0
            fi
            start_services "native" "" "$synthesia_key" "$unsplash_key"
            ;;
        2)
            start_services "docker" "" "$synthesia_key" "$unsplash_key"
            ;;
        3)
            echo -e "${CYAN}Using Ollama Cloud.${NC}"
            local ollama_key
            ollama_key=$(prompt_api_key_simple) || exit 0
            start_services "cloud" "$ollama_key" "$synthesia_key" "$unsplash_key"
            ;;
        4)
            echo -e "${CYAN}Stopping SecuSense...${NC}"
            docker compose --profile with-ollama down
            echo -e "${GREEN}Stopped.${NC}"
            ;;
        5|*)
            exit 0
            ;;
    esac
}

# Main
if ! docker compose version &>/dev/null 2>&1; then
    echo -e "${RED}Error: docker compose not found${NC}"
    exit 1
fi

show_tui_menu || show_simple_menu
