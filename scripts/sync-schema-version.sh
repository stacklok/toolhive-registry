#!/bin/bash
set -euo pipefail

# sync-schema-version.sh
# Syncs the schema reference commit SHA with the Go module version
# Ensures schema validation uses the exact same version as the Go code

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
SCHEMA_FILE="$PROJECT_ROOT/schemas/registry.schema.json"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log() {
    echo -e "${BLUE}[sync-schema]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[sync-schema]${NC} $1"
}

error() {
    echo -e "${RED}[sync-schema]${NC} $1" >&2
}

success() {
    echo -e "${GREEN}[sync-schema]${NC} $1"
}

# Check if jq is available
check_dependencies() {
    if ! command -v jq >/dev/null 2>&1; then
        error "jq is required but not installed. Please install it:"
        error "  macOS: brew install jq"
        error "  Ubuntu: apt-get install jq"
        exit 1
    fi
}

# Function to extract commit SHA from go.mod version
get_current_commit_sha() {
    local version
    version=$(grep "github.com/modelcontextprotocol/registry" "$PROJECT_ROOT/go.mod" | awk '{print $2}')
    
    if [[ "$version" =~ v[0-9]+\.[0-9]+\.[0-9]+-[0-9]+-([a-f0-9]+)$ ]]; then
        # Extract SHA from pseudo-version (v0.0.0-20250903150202-6ea3828e3ce6)
        echo "${BASH_REMATCH[1]}"
    elif [[ "$version" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        # For tagged versions, we need to resolve to commit SHA
        warn "Tagged version detected: $version" >&2
        warn "Will attempt to resolve commit SHA from GitHub..." >&2
        
        # Try to get commit SHA for the tag
        local sha
        sha=$(curl -s "https://api.github.com/repos/modelcontextprotocol/registry/git/refs/tags/$version" | \
              jq -r '.object.sha // empty' 2>/dev/null)
        
        if [[ -n "$sha" ]]; then
            echo "$sha"
        else
            error "Failed to resolve commit SHA for tagged version: $version"
            return 1
        fi
    else
        error "Unable to parse version format: $version"
        return 1
    fi
}

# Function to get current SHA from schema
get_schema_commit_sha() {
    if [[ ! -f "$SCHEMA_FILE" ]]; then
        error "Schema file not found: $SCHEMA_FILE"
        return 1
    fi
    
    # Extract SHA from the GitHub raw URL using jq
    local ref_url
    ref_url=$(jq -r '.properties.data.properties.servers.items["$ref"] // empty' "$SCHEMA_FILE")
    
    if [[ -n "$ref_url" ]]; then
        # Extract SHA from URL like: https://raw.githubusercontent.com/.../registry/6ea3828e3ce62cfd9815376cd6825453da011fa1/docs/...
        echo "$ref_url" | sed 's|.*/registry/\([^/]*\)/.*|\1|'
    else
        error "No schema reference URL found"
        return 1
    fi
}

# Function to get Go module version for metadata
get_go_module_version() {
    grep "github.com/modelcontextprotocol/registry" "$PROJECT_ROOT/go.mod" | awk '{print $2}'
}

# Function to update schema with new commit SHA using jq
update_schema_reference() {
    local new_sha="$1"
    local old_sha="$2"
    local go_version="$3"
    
    log "Updating schema reference from $old_sha to $new_sha"
    
    
    local timestamp
    timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    
    # Update both the $ref URL and add/update metadata using jq
    jq --arg new_sha "$new_sha" \
       --arg old_sha "$old_sha" \
       --arg go_version "$go_version" \
       --arg timestamp "$timestamp" \
       '
       # Update the $ref URL
       .properties.data.properties.servers.items["$ref"] |= 
         gsub("/registry/[^/]+/"; 
              "/registry/" + $new_sha + "/") |
       
       # Add or update _schema_version metadata
       ._schema_version = {
         "mcp_registry_version": $go_version,
         "mcp_registry_commit": $new_sha,
         "updated_at": $timestamp,
         "updated_by": "sync-schema-version.sh"
       }
       ' "$SCHEMA_FILE" > "$SCHEMA_FILE.tmp"
    
    # Replace original with updated version
    mv "$SCHEMA_FILE.tmp" "$SCHEMA_FILE"
    
    success "Schema reference updated to commit $new_sha"
}


# Function to show what changed
show_changes() {
    local new_sha="$1"
    local old_sha="$2"
    
    log "Changes summary:"
    log "  Schema reference updated:"
    log "    From: https://raw.githubusercontent.com/.../registry/$old_sha/docs/..."
    log "    To:   https://raw.githubusercontent.com/.../registry/$new_sha/docs/..."
    log ""
    
    if command -v git >/dev/null 2>&1 && git rev-parse --git-dir >/dev/null 2>&1; then
        log "Git diff:"
        git --no-pager diff "$SCHEMA_FILE" || true
    fi
}

# Main function
main() {
    log "Starting schema version sync..."
    
    # Check dependencies first
    check_dependencies
    
    cd "$PROJECT_ROOT"
    
    # Get current commit SHA from go.mod
    log "Extracting commit SHA from go.mod..."
    local current_sha
    if ! current_sha=$(get_current_commit_sha); then
        error "Failed to extract commit SHA from go.mod"
        exit 1
    fi
    
    # Get Go module version for metadata
    local go_version
    go_version=$(get_go_module_version)
    
    log "Current Go module version: $go_version"
    log "Current Go module commit SHA: $current_sha"
    
    # Get current SHA from schema
    log "Extracting commit SHA from schema..."
    local schema_sha
    if ! schema_sha=$(get_schema_commit_sha); then
        error "Failed to extract commit SHA from schema"
        exit 1
    fi
    
    log "Current schema commit SHA: $schema_sha"
    
    # Compare and update if different (handle short vs long SHA)
    if [[ "$schema_sha" == "$current_sha"* ]] || [[ "$current_sha" == "$schema_sha"* ]]; then
        success "Schema reference is already in sync! (SHA: $current_sha)"
        exit 0
    fi
    
    warn "Schema reference is out of sync!"
    warn "  Go module SHA: $current_sha"
    warn "  Schema SHA:    $schema_sha"
    
    # Update schema reference
    if ! update_schema_reference "$current_sha" "$schema_sha" "$go_version"; then
        error "Failed to update schema reference"
        exit 1
    fi
    
    # Show what changed  
    show_changes "$current_sha" "$schema_sha"
    
    success "Schema sync completed successfully!"
    log ""
    log "Next steps:"
    log "  1. Review the changes above"
    log "  2. Test the registry build: task build:registry"
    log "  3. Commit the changes if everything looks good"
}

# Run main function
main "$@"