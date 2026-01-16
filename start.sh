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

    echo -e "${CYAN}Starting SecuSense...${NC}"

    case $ollama_mode in
        docker)
            echo -e "${BLUE}Starting with Docker Ollama${NC}"
            OLLAMA_URL="http://ollama:11434" docker compose --profile with-ollama up -d
            ;;
        cloud)
            echo -e "${BLUE}Using Ollama Cloud${NC}"
            OLLAMA_URL="https://api.ollama.com" docker compose up -d
            ;;
        *)
            echo -e "${BLUE}Using native Ollama (host.docker.internal:11434)${NC}"
            docker compose up -d
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

    case $choice in
        native)
            if ! check_native_ollama; then
                $cmd --title "Warning" --yesno "Native Ollama not detected.\n\nMake sure Ollama is running:\n  ollama serve\n\nContinue anyway?" 12 50 || exit 0
            fi
            start_services "native"
            ;;
        docker)
            start_services "docker"
            ;;
        cloud)
            $cmd --title "Ollama Cloud" --msgbox "Using Ollama Cloud.\n\nMake sure you have:\n1. An Ollama account at ollama.com\n2. Run 'ollama login' locally first to authenticate\n\nCloud models will be fetched on-demand." 14 55
            start_services "cloud"
            ;;
        stop)
            echo -e "${CYAN}Stopping SecuSense...${NC}"
            docker compose --profile with-ollama down
            echo -e "${GREEN}Stopped.${NC}"
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

    case $choice in
        1)
            if ! check_native_ollama; then
                echo -e "${RED}Warning: Native Ollama not detected.${NC}"
                read -p "Continue anyway? [y/N]: " confirm
                [[ "$confirm" =~ ^[Yy]$ ]] || exit 0
            fi
            start_services "native"
            ;;
        2)
            start_services "docker"
            ;;
        3)
            echo -e "${CYAN}Using Ollama Cloud.${NC}"
            echo ""
            echo "Make sure you have:"
            echo "  1. An Ollama account at ollama.com"
            echo "  2. Run 'ollama login' locally first to authenticate"
            echo ""
            start_services "cloud"
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
