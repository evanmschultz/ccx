#!/usr/bin/env bash

# ccx - Claude Code eXchange
# Multi-account switcher with beautiful gum-powered interface
# Works with bash 3.2+ (macOS default)

set -euo pipefail

# Version info
readonly VERSION="1.0.0"
readonly UPDATE_URL="https://raw.githubusercontent.com/yourusername/ccx/main/ccx.sh"
readonly VERSION_URL="https://raw.githubusercontent.com/yourusername/ccx/main/VERSION"

# Configuration paths
readonly CCX_DIR="$HOME/.ccx"
readonly BACKUP_DIR="$CCX_DIR/accounts"
readonly STATE_FILE="$CCX_DIR/state.json"
readonly UPDATE_CHECK_FILE="$CCX_DIR/.update_check"

# Platform detection
detect_platform() {
    case "$(uname -s)" in
        Darwin) echo "macos" ;;
        Linux) 
            if [[ -n "${WSL_DISTRO_NAME:-}" ]]; then
                echo "wsl"
            else
                echo "linux"
            fi
            ;;
        *) echo "unknown" ;;
    esac
}

# Check if running in container
is_running_in_container() {
    [[ -f /.dockerenv ]] || \
    [[ -f /proc/1/cgroup ]] && grep -q 'docker\|lxc\|containerd\|kubepods' /proc/1/cgroup 2>/dev/null || \
    [[ -n "${CONTAINER:-}" ]] || [[ -n "${container:-}" ]]
}

# Check dependencies
check_dependencies() {
    local missing=()
    
    for cmd in jq gum; do
        if ! command -v "$cmd" >/dev/null 2>&1; then
            missing+=("$cmd")
        fi
    done
    
    if [[ ${#missing[@]} -gt 0 ]]; then
        echo "Error: Missing required dependencies: ${missing[*]}"
        echo ""
        echo "Install with:"
        echo "  macOS:  brew install ${missing[*]}"
        echo "  Linux:  apt install ${missing[*]} (or equivalent)"
        exit 1
    fi
}

# Setup directories
setup_directories() {
    mkdir -p "$CCX_DIR" "$BACKUP_DIR"
    chmod 700 "$CCX_DIR" "$BACKUP_DIR"
}

# Initialize state file
init_state_file() {
    if [[ ! -f "$STATE_FILE" ]]; then
        cat > "$STATE_FILE" <<EOF
{
  "version": "$VERSION",
  "accounts": {},
  "activeAccount": null,
  "history": [],
  "lastUpdated": "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
}
EOF
        chmod 600 "$STATE_FILE"
    fi
}

# Get Claude config path
get_claude_config_path() {
    local primary="$HOME/.claude/.claude.json"
    local fallback="$HOME/.claude.json"
    
    if [[ -f "$primary" ]] && jq -e '.oauthAccount' "$primary" >/dev/null 2>&1; then
        echo "$primary"
    else
        echo "$fallback"
    fi
}

# Safe JSON write
write_json() {
    local file="$1"
    local content="$2"
    local temp_file
    temp_file=$(mktemp)
    
    echo "$content" > "$temp_file"
    if jq . "$temp_file" >/dev/null 2>&1; then
        mv "$temp_file" "$file"
        chmod 600 "$file"
    else
        rm -f "$temp_file"
        return 1
    fi
}

# Check for updates (non-intrusive)
check_for_updates() {
    local last_check=0
    local current_time
    current_time=$(date +%s)
    
    if [[ -f "$UPDATE_CHECK_FILE" ]]; then
        last_check=$(cat "$UPDATE_CHECK_FILE" 2>/dev/null || echo 0)
    fi
    
    # Check once per day
    if (( current_time - last_check > 86400 )); then
        echo "$current_time" > "$UPDATE_CHECK_FILE"
        
        # Check version in background
        (
            if command -v curl >/dev/null 2>&1; then
                remote_version=$(curl -sL "$VERSION_URL" 2>/dev/null || echo "")
                if [[ -n "$remote_version" ]] && [[ "$remote_version" != "$VERSION" ]]; then
                    echo "$remote_version" > "$CCX_DIR/.update_available"
                fi
            fi
        ) &
    fi
    
    # Show update notice if available
    if [[ -f "$CCX_DIR/.update_available" ]]; then
        local available_version
        available_version=$(cat "$CCX_DIR/.update_available")
        export CCX_UPDATE_AVAILABLE="$available_version"
    fi
}

# Get current account info
get_current_account() {
    local config_path
    config_path=$(get_claude_config_path)
    
    if [[ ! -f "$config_path" ]]; then
        echo ""
        return
    fi
    
    jq -r '.oauthAccount.emailAddress // empty' "$config_path" 2>/dev/null || echo ""
}

# Get account display name (with alias if exists)
get_account_display() {
    local account_id="$1"
    local state_data
    state_data=$(cat "$STATE_FILE")
    
    local email alias
    email=$(echo "$state_data" | jq -r --arg id "$account_id" '.accounts[$id].email // ""')
    alias=$(echo "$state_data" | jq -r --arg id "$account_id" '.accounts[$id].alias // ""')
    
    if [[ -n "$alias" ]]; then
        echo "$alias ($email)"
    else
        echo "$email"
    fi
}

# Read credentials (platform-specific)
read_credentials() {
    local platform
    platform=$(detect_platform)
    
    case "$platform" in
        macos)
            security find-generic-password -s "Claude Code-credentials" -w 2>/dev/null || echo ""
            ;;
        linux|wsl)
            [[ -f "$HOME/.claude/.credentials.json" ]] && cat "$HOME/.claude/.credentials.json" || echo ""
            ;;
    esac
}

# Write credentials (platform-specific)
write_credentials() {
    local credentials="$1"
    local platform
    platform=$(detect_platform)
    
    case "$platform" in
        macos)
            security add-generic-password -U -s "Claude Code-credentials" -a "$USER" -w "$credentials" 2>/dev/null
            ;;
        linux|wsl)
            mkdir -p "$HOME/.claude"
            printf '%s' "$credentials" > "$HOME/.claude/.credentials.json"
            chmod 600 "$HOME/.claude/.credentials.json"
            ;;
    esac
}

# Save account backup
save_account_backup() {
    local account_id="$1"
    local email="$2"
    local credentials="$3"
    local config="$4"
    local platform
    platform=$(detect_platform)
    
    # Save config
    echo "$config" > "$BACKUP_DIR/${account_id}.config.json"
    chmod 600 "$BACKUP_DIR/${account_id}.config.json"
    
    # Save credentials
    case "$platform" in
        macos)
            security add-generic-password -U -s "ccx-account-${account_id}" -a "$email" -w "$credentials" 2>/dev/null
            ;;
        linux|wsl)
            printf '%s' "$credentials" > "$BACKUP_DIR/${account_id}.credentials.json"
            chmod 600 "$BACKUP_DIR/${account_id}.credentials.json"
            ;;
    esac
}

# Load account backup
load_account_backup() {
    local account_id="$1"
    local email="$2"
    local platform
    platform=$(detect_platform)
    
    local config credentials
    
    # Load config
    if [[ -f "$BACKUP_DIR/${account_id}.config.json" ]]; then
        config=$(cat "$BACKUP_DIR/${account_id}.config.json")
    else
        return 1
    fi
    
    # Load credentials
    case "$platform" in
        macos)
            credentials=$(security find-generic-password -s "ccx-account-${account_id}" -w 2>/dev/null || echo "")
            ;;
        linux|wsl)
            if [[ -f "$BACKUP_DIR/${account_id}.credentials.json" ]]; then
                credentials=$(cat "$BACKUP_DIR/${account_id}.credentials.json")
            fi
            ;;
    esac
    
    if [[ -z "$credentials" ]]; then
        return 1
    fi
    
    echo "$credentials|CCX_SEPARATOR|$config"
}

# Add current account
add_current_account() {
    local current_email
    current_email=$(get_current_account)
    
    if [[ -z "$current_email" ]]; then
        gum style --foreground 196 "No active Claude account found. Please log in first."
        return 1
    fi
    
    # Check if already exists
    local exists
    exists=$(jq -r --arg email "$current_email" '[.accounts[] | select(.email == $email)] | length' "$STATE_FILE")
    
    if [[ "$exists" -gt 0 ]]; then
        gum style --foreground 226 "Account $current_email is already managed."
        return 0
    fi
    
    # Get credentials and config
    local credentials config
    credentials=$(read_credentials)
    config=$(cat "$(get_claude_config_path)")
    
    if [[ -z "$credentials" ]]; then
        gum style --foreground 196 "Error: No credentials found for current account"
        return 1
    fi
    
    # Generate account ID
    local account_id
    account_id=$(uuidgen | tr '[:upper:]' '[:lower:]' | tr -d '-' | cut -c1-8)
    
    # Ask for alias (optional)
    local alias=""
    if gum confirm "Would you like to set an alias for this account?"; then
        alias=$(gum input --placeholder "Enter alias (e.g., 'work', 'personal')")
    fi
    
    # Save backup
    save_account_backup "$account_id" "$current_email" "$credentials" "$config"
    
    # Update state
    local updated_state
    updated_state=$(jq --arg id "$account_id" \
                       --arg email "$current_email" \
                       --arg alias "$alias" \
                       --arg uuid "$(echo "$config" | jq -r '.oauthAccount.accountUuid')" \
                       --arg now "$(date -u +%Y-%m-%dT%H:%M:%SZ)" '
        .accounts[$id] = {
            email: $email,
            alias: (if $alias == "" then null else $alias end),
            uuid: $uuid,
            added: $now,
            lastUsed: $now
        } |
        .activeAccount = $id |
        .lastUpdated = $now
    ' "$STATE_FILE")
    
    write_json "$STATE_FILE" "$updated_state"
    
    gum style --foreground 82 "✓ Added account: $(get_account_display "$account_id")"
}

# Switch to account
switch_to_account() {
    local target_id="$1"
    local state_data
    state_data=$(cat "$STATE_FILE")
    
    local target_email
    target_email=$(echo "$state_data" | jq -r --arg id "$target_id" '.accounts[$id].email // ""')
    
    if [[ -z "$target_email" ]]; then
        gum style --foreground 196 "Error: Account not found"
        return 1
    fi
    
    # Check if Claude is running
    if pgrep -x "claude" >/dev/null 2>&1; then
        gum style --foreground 226 "Claude Code is running. Please close it first."
        gum spin --spinner dot --title "Waiting for Claude Code to close..." -- bash -c 'while pgrep -x "claude" >/dev/null 2>&1; do sleep 1; done'
    fi
    
    # Get current account
    local current_email current_id
    current_email=$(get_current_account)
    current_id=$(echo "$state_data" | jq -r --arg email "$current_email" '.accounts | to_entries[] | select(.value.email == $email) | .key // ""')
    
    # Backup current if it's managed
    if [[ -n "$current_id" ]]; then
        local current_creds current_config
        current_creds=$(read_credentials)
        current_config=$(cat "$(get_claude_config_path)")
        
        if [[ -n "$current_creds" ]] && [[ -n "$current_config" ]]; then
            save_account_backup "$current_id" "$current_email" "$current_creds" "$current_config"
        fi
    fi
    
    # Load target account
    local backup_data
    backup_data=$(load_account_backup "$target_id" "$target_email")
    
    if [[ -z "$backup_data" ]]; then
        gum style --foreground 196 "Error: Failed to load account backup"
        return 1
    fi
    
    # Split credentials and config
    local credentials config
    credentials="${backup_data%%|CCX_SEPARATOR|*}"
    config="${backup_data#*|CCX_SEPARATOR|}"
    
    # Apply credentials and config
    write_credentials "$credentials"
    
    # Merge config
    local current_full_config merged_config
    current_full_config=$(cat "$(get_claude_config_path)")
    merged_config=$(echo "$current_full_config" | jq --argjson oauth "$(echo "$config" | jq '.oauthAccount')" '.oauthAccount = $oauth')
    
    write_json "$(get_claude_config_path)" "$merged_config"
    
    # Update state and history
    local updated_state
    updated_state=$(echo "$state_data" | jq --arg id "$target_id" \
                                            --arg now "$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
                                            --arg from "$current_email" \
                                            --arg to "$target_email" '
        .activeAccount = $id |
        .accounts[$id].lastUsed = $now |
        .history = ([{
            from: $from,
            to: $to,
            timestamp: $now
        }] + .history) | .[0:10] |
        .lastUpdated = $now
    ')
    
    write_json "$STATE_FILE" "$updated_state"
    
    # Update shell environment
    export CCX_CURRENT_ACCOUNT="$target_email"
    
    gum style --foreground 82 "✓ Switched to: $(get_account_display "$target_id")"
    gum style --foreground 245 "Please restart Claude Code to use the new account."
}

# Menu interface
show_menu() {
    local state_data
    state_data=$(cat "$STATE_FILE")
    
    # Build menu options
    local options=()
    local account_ids=()
    
    # Add account options
    while IFS= read -r line; do
        local id email alias active
        id=$(echo "$line" | cut -d'|' -f1)
        email=$(echo "$line" | cut -d'|' -f2)
        alias=$(echo "$line" | cut -d'|' -f3)
        active=$(echo "$line" | cut -d'|' -f4)
        
        local display
        if [[ -n "$alias" ]] && [[ "$alias" != "null" ]]; then
            display="$alias ($email)"
        else
            display="$email"
        fi
        
        if [[ "$active" == "true" ]]; then
            display="$display ✓"
        fi
        
        options+=("Switch to: $display")
        account_ids+=("$id")
    done < <(echo "$state_data" | jq -r --arg active "$(echo "$state_data" | jq -r '.activeAccount')" '
        .accounts | to_entries[] | 
        "\(.key)|\(.value.email)|\(.value.alias // "")|\(.key == $active)"
    ')
    
    # Add management options
    options+=("Add current account" "Manage aliases" "Remove account" "View history" "Check for updates" "Exit")
    
    # Show update notice if available
    local title="ccx - Claude Code eXchange"
    if [[ -n "${CCX_UPDATE_AVAILABLE:-}" ]]; then
        title="$title (update available: v${CCX_UPDATE_AVAILABLE})"
    fi
    
    # Show menu
    local choice
    choice=$(printf '%s\n' "${options[@]}" | gum choose --header "$title")
    
    case "$choice" in
        "Switch to: "*)
            # Extract account index
            for i in "${!options[@]}"; do
                if [[ "${options[$i]}" == "$choice" ]]; then
                    switch_to_account "${account_ids[$i]}"
                    break
                fi
            done
            ;;
        "Add current account")
            add_current_account
            ;;
        "Manage aliases")
            manage_aliases
            ;;
        "Remove account")
            remove_account_interactive
            ;;
        "View history")
            view_history
            ;;
        "Check for updates")
            check_and_update
            ;;
        "Exit")
            exit 0
            ;;
    esac
}

# Manage aliases
manage_aliases() {
    local state_data
    state_data=$(cat "$STATE_FILE")
    
    # Build account list
    local accounts=()
    local account_ids=()
    
    while IFS= read -r line; do
        local id email alias
        id=$(echo "$line" | cut -d'|' -f1)
        email=$(echo "$line" | cut -d'|' -f2)
        alias=$(echo "$line" | cut -d'|' -f3)
        
        if [[ -n "$alias" ]] && [[ "$alias" != "null" ]]; then
            accounts+=("$email (current alias: $alias)")
        else
            accounts+=("$email (no alias)")
        fi
        account_ids+=("$id")
    done < <(echo "$state_data" | jq -r '.accounts | to_entries[] | "\(.key)|\(.value.email)|\(.value.alias // "")"')
    
    if [[ ${#accounts[@]} -eq 0 ]]; then
        gum style --foreground 226 "No accounts to manage."
        return
    fi
    
    # Select account
    local choice
    choice=$(printf '%s\n' "${accounts[@]}" | gum choose --header "Select account to set alias:")
    
    # Find selected account
    for i in "${!accounts[@]}"; do
        if [[ "${accounts[$i]}" == "$choice" ]]; then
            local new_alias
            new_alias=$(gum input --placeholder "Enter new alias (leave empty to remove)")
            
            # Update alias
            local updated_state
            updated_state=$(echo "$state_data" | jq --arg id "${account_ids[$i]}" \
                                                    --arg alias "$new_alias" '
                .accounts[$id].alias = (if $alias == "" then null else $alias end)
            ')
            
            write_json "$STATE_FILE" "$updated_state"
            gum style --foreground 82 "✓ Alias updated"
            break
        fi
    done
}

# Remove account interactive
remove_account_interactive() {
    local state_data
    state_data=$(cat "$STATE_FILE")
    
    # Build account list
    local accounts=()
    local account_ids=()
    
    while IFS= read -r line; do
        local id display
        id=$(echo "$line" | cut -d'|' -f1)
        display=$(get_account_display "$id")
        accounts+=("$display")
        account_ids+=("$id")
    done < <(echo "$state_data" | jq -r '.accounts | to_entries[] | "\(.key)"')
    
    if [[ ${#accounts[@]} -eq 0 ]]; then
        gum style --foreground 226 "No accounts to remove."
        return
    fi
    
    # Select account
    local choice
    choice=$(printf '%s\n' "${accounts[@]}" | gum choose --header "Select account to remove:")
    
    # Find selected account
    for i in "${!accounts[@]}"; do
        if [[ "${accounts[$i]}" == "$choice" ]]; then
            if gum confirm "Are you sure you want to remove this account?"; then
                remove_account "${account_ids[$i]}"
            fi
            break
        fi
    done
}

# Remove account
remove_account() {
    local account_id="$1"
    local platform
    platform=$(detect_platform)
    
    # Remove backups
    rm -f "$BACKUP_DIR/${account_id}.config.json"
    
    case "$platform" in
        macos)
            security delete-generic-password -s "ccx-account-${account_id}" 2>/dev/null || true
            ;;
        linux|wsl)
            rm -f "$BACKUP_DIR/${account_id}.credentials.json"
            ;;
    esac
    
    # Update state
    local state_data updated_state
    state_data=$(cat "$STATE_FILE")
    updated_state=$(echo "$state_data" | jq --arg id "$account_id" 'del(.accounts[$id])')
    
    write_json "$STATE_FILE" "$updated_state"
    
    gum style --foreground 82 "✓ Account removed"
}

# View history
view_history() {
    local state_data
    state_data=$(cat "$STATE_FILE")
    
    local history
    history=$(echo "$state_data" | jq -r '.history[] | "\(.timestamp | split("T")[0]) \(.timestamp | split("T")[1] | split(".")[0]): \(.from) → \(.to)"')
    
    if [[ -z "$history" ]]; then
        gum style --foreground 226 "No switch history yet."
    else
        echo "$history" | gum pager --help="Switch History"
    fi
}

# Check and update
check_and_update() {
    gum spin --spinner dot --title "Checking for updates..." -- sleep 1
    
    local remote_version
    if command -v curl >/dev/null 2>&1; then
        remote_version=$(curl -sL "$VERSION_URL" 2>/dev/null || echo "")
    fi
    
    if [[ -z "$remote_version" ]]; then
        gum style --foreground 196 "Failed to check for updates"
        return
    fi
    
    if [[ "$remote_version" == "$VERSION" ]]; then
        gum style --foreground 82 "You're running the latest version ($VERSION)"
        return
    fi
    
    gum style --foreground 226 "New version available: $remote_version (current: $VERSION)"
    
    if gum confirm "Would you like to update now?"; then
        local temp_file
        temp_file=$(mktemp)
        
        if curl -fsSL "$UPDATE_URL" -o "$temp_file"; then
            if [[ -s "$temp_file" ]]; then
                cp "$0" "$0.bak"
                mv "$temp_file" "$0"
                chmod +x "$0"
                rm -f "$CCX_DIR/.update_available"
                gum style --foreground 82 "✓ Updated to version $remote_version"
                gum style --foreground 245 "Please run 'ccx' again to use the new version"
                exit 0
            fi
        fi
        
        rm -f "$temp_file"
        gum style --foreground 196 "Update failed. Please update manually."
    fi
}

# Quick switch by identifier
quick_switch() {
    local identifier="$1"
    local state_data
    state_data=$(cat "$STATE_FILE")
    
    local target_id=""
    
    # Try to match by: account ID, email, alias, or index
    if echo "$state_data" | jq -e --arg id "$identifier" '.accounts[$id]' >/dev/null 2>&1; then
        # Direct ID match
        target_id="$identifier"
    else
        # Try email match
        target_id=$(echo "$state_data" | jq -r --arg email "$identifier" '
            .accounts | to_entries[] | select(.value.email == $email) | .key // ""
        ' | head -n1)
        
        # Try alias match
        if [[ -z "$target_id" ]]; then
            target_id=$(echo "$state_data" | jq -r --arg alias "$identifier" '
                .accounts | to_entries[] | select(.value.alias == $alias) | .key // ""
            ' | head -n1)
        fi
        
        # Try numeric index
        if [[ -z "$target_id" ]] && [[ "$identifier" =~ ^[0-9]+$ ]]; then
            local index=$((identifier - 1))
            target_id=$(echo "$state_data" | jq -r --arg idx "$index" '
                .accounts | to_entries | .[$idx | tonumber].key // ""
            ')
        fi
    fi
    
    if [[ -n "$target_id" ]]; then
        switch_to_account "$target_id"
    else
        gum style --foreground 196 "No account found matching: $identifier"
        exit 1
    fi
}

# Export function for shell integration
export_shell_function() {
    cat <<'EOF'
# ccx shell integration
ccx_prompt() {
    if [[ -f "$HOME/.ccx/state.json" ]]; then
        local current_account=$(jq -r '.activeAccount as $id | .accounts[$id] | .alias // .email // ""' "$HOME/.ccx/state.json" 2>/dev/null)
        if [[ -n "$current_account" ]]; then
            echo " [$current_account]"
        fi
    fi
}

# Add to your PS1:
# PS1="${PS1}$(ccx_prompt) "
EOF
}

# Main function
main() {
    # Don't run as root unless in container
    if [[ $EUID -eq 0 ]] && ! is_running_in_container; then
        gum style --foreground 196 "Error: Do not run as root"
        exit 1
    fi
    
    check_dependencies
    setup_directories
    init_state_file
    check_for_updates
    
    case "${1:-}" in
        # Quick switch
        [0-9]*|*@*)
            quick_switch "$1"
            ;;
        # CLI commands
        add|--add)
            add_current_account
            ;;
        list|--list|-l)
            local state_data
            state_data=$(cat "$STATE_FILE")
            echo "$state_data" | jq -r '.accounts | to_entries[] | 
                "\(.key): \(.value.alias // .value.email) \(if .value.email != (.value.alias // "") then "(" + .value.email + ")" else "" end)"'
            ;;
        switch|--switch|-s)
            if [[ -n "${2:-}" ]]; then
                quick_switch "$2"
            else
                # Switch to next in sequence
                local state_data current next_id
                state_data=$(cat "$STATE_FILE")
                current=$(echo "$state_data" | jq -r '.activeAccount // ""')
                
                if [[ -n "$current" ]]; then
                    # Get all account IDs
                    local ids=()
                    while IFS= read -r id; do
                        ids+=("$id")
                    done < <(echo "$state_data" | jq -r '.accounts | keys[]')
                    
                    # Find current index and get next
                    for i in "${!ids[@]}"; do
                        if [[ "${ids[$i]}" == "$current" ]]; then
                            next_id="${ids[$(( (i + 1) % ${#ids[@]} ))]}"
                            break
                        fi
                    done
                    
                    if [[ -n "$next_id" ]]; then
                        switch_to_account "$next_id"
                    fi
                else
                    gum style --foreground 196 "No active account"
                fi
            fi
            ;;
        remove|--remove|-r)
            if [[ -n "${2:-}" ]]; then
                # Find account by identifier
                local state_data target_id
                state_data=$(cat "$STATE_FILE")
                
                # Similar logic to quick_switch for finding account
                # ... (implement same search logic)
                
                if [[ -n "$target_id" ]]; then
                    if gum confirm "Remove account $(get_account_display "$target_id")?"; then
                        remove_account "$target_id"
                    fi
                else
                    gum style --foreground 196 "Account not found: $2"
                fi
            else
                remove_account_interactive
            fi
            ;;
        history|--history|-h)
            view_history
            ;;
        update|--update|-u)
            check_and_update
            ;;
        shell-integration|--shell-integration)
            export_shell_function
            ;;
        version|--version|-v)
            echo "ccx version $VERSION"
            if [[ -n "${CCX_UPDATE_AVAILABLE:-}" ]]; then
                echo "Update available: v${CCX_UPDATE_AVAILABLE}"
            fi
            ;;
        help|--help|-h)
            cat <<EOF
ccx - Claude Code eXchange
Version $VERSION

Usage:
  ccx                      Interactive menu mode
  ccx <number>            Quick switch by index (1, 2, 3...)
  ccx <email>             Quick switch by email
  ccx <alias>             Quick switch by alias

Commands:
  add, --add              Add current Claude account
  list, --list, -l        List all accounts
  switch, --switch, -s    Switch to next account (or specific with argument)
  remove, --remove, -r    Remove account (interactive or by identifier)
  history, --history, -h  View switch history
  update, --update, -u    Check for and install updates
  shell-integration       Export shell prompt function
  version, --version, -v  Show version info
  help, --help           Show this help

Examples:
  ccx                     # Open interactive menu
  ccx 2                   # Switch to account #2
  ccx work                # Switch to account with alias 'work'
  ccx user@example.com    # Switch to account by email
  ccx add                 # Add current account
  ccx switch work         # Switch to 'work' account

Shell Integration:
  eval "\$(ccx shell-integration)"
  # Then add \$(ccx_prompt) to your PS1
EOF
            ;;
        "")
            # No arguments - show menu
            show_menu
            ;;
        *)
            # Try as quick switch identifier
            quick_switch "$1"
            ;;
    esac
}

# Run main function
main "$@"