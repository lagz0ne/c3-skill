#!/bin/bash
#
# c3-init.sh - Initialize C3 architecture documentation structure
#
# Usage:
#   PROJECT="MyApp" C1="backend" C2="frontend" ./c3-init.sh
#
# Environment variables:
#   PROJECT     - Project name (required)
#   SUMMARY     - One-line project summary (optional)
#   C1, C2, ... - Container slugs (at least C1 required)
#   C1_NAME, C2_NAME, ... - Container display names (optional, defaults to slug)
#

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Resolve script directory to find templates
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TEMPLATE_DIR="${SCRIPT_DIR}/../templates"

# Validate required inputs
if [ -z "$PROJECT" ]; then
    echo -e "${RED}Error: PROJECT is required${NC}"
    echo "Usage: PROJECT=\"MyApp\" C1=\"backend\" ./c3-init.sh"
    exit 1
fi

if [ -z "$C1" ]; then
    echo -e "${RED}Error: At least C1 (first container) is required${NC}"
    echo "Usage: PROJECT=\"MyApp\" C1=\"backend\" ./c3-init.sh"
    exit 1
fi

# Defaults
SUMMARY="${SUMMARY:-One-line summary of ${PROJECT}}"
DATE=$(date +%Y-%m-%d)

# Check if .c3 already exists
if [ -d ".c3" ]; then
    echo -e "${YELLOW}Warning: .c3/ directory already exists${NC}"
    read -p "Overwrite? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Aborted."
        exit 1
    fi
fi

echo -e "${GREEN}Initializing C3 structure for ${PROJECT}...${NC}"

# Create directory structure
mkdir -p .c3/adr

# Count containers and create directories
CONTAINER_COUNT=0
for i in {1..9}; do
    VAR="C${i}"
    if [ -n "${!VAR}" ]; then
        CONTAINER_COUNT=$i
        SLUG="${!VAR}"
        mkdir -p ".c3/c3-${i}-${SLUG}"
        echo "  Created .c3/c3-${i}-${SLUG}/"
    fi
done

# Function to substitute variables in template
# Falls back to sed if envsubst is not available
substitute_template() {
    local template="$1"
    local output="$2"

    if command -v envsubst &> /dev/null; then
        # Use envsubst (preferred - handles all env vars)
        envsubst < "$template" > "$output"
    else
        # Fallback: use sed for known variables
        # Note: Order matters - replace ${N}${NN} patterns before ${N}
        sed -e "s|\${PROJECT}|${PROJECT:-}|g" \
            -e "s|\${SUMMARY}|${SUMMARY:-}|g" \
            -e "s|\${DATE}|${DATE:-}|g" \
            -e "s|\${CONTAINER_NAME}|${CONTAINER_NAME:-}|g" \
            -e "s|\${COMPONENT_NAME}|${COMPONENT_NAME:-}|g" \
            -e "s|\${NN}|${NN:-}|g" \
            -e "s|\${N}|${N:-}|g" \
            -e "s|\${C1_NAME}|${C1_NAME:-}|g" \
            -e "s|\${C2_NAME}|${C2_NAME:-}|g" \
            -e "s|\${C3_NAME}|${C3_NAME:-}|g" \
            < "$template" > "$output"
    fi
}

# Create context (README.md)
echo "  Creating .c3/README.md (Context)"
export PROJECT SUMMARY DATE
export C1_NAME="${C1_NAME:-${C1}}"
substitute_template "${TEMPLATE_DIR}/context.md" ".c3/README.md"

# Create container READMEs
for i in $(seq 1 $CONTAINER_COUNT); do
    VAR="C${i}"
    NAME_VAR="C${i}_NAME"
    SLUG="${!VAR}"

    export N="$i"
    export CONTAINER_NAME="${!NAME_VAR:-${SLUG}}"
    export SUMMARY="Container ${i} summary"

    echo "  Creating .c3/c3-${i}-${SLUG}/README.md (Container)"
    substitute_template "${TEMPLATE_DIR}/container.md" ".c3/c3-${i}-${SLUG}/README.md"
done

# Create ADR-000
echo "  Creating .c3/adr/adr-00000000-c3-adoption.md (ADR)"
export PROJECT DATE
substitute_template "${TEMPLATE_DIR}/adr-000.md" ".c3/adr/adr-00000000-c3-adoption.md"

# Summary
echo ""
echo -e "${GREEN}C3 structure created:${NC}"
echo "  .c3/"
echo "  ├── README.md                         (Context - c3-0)"
for i in $(seq 1 $CONTAINER_COUNT); do
    VAR="C${i}"
    SLUG="${!VAR}"
    echo "  ├── c3-${i}-${SLUG}/"
    echo "  │   └── README.md                     (Container - c3-${i})"
done
echo "  └── adr/"
echo "      └── adr-00000000-c3-adoption.md   (Adoption ADR)"
echo ""
echo -e "${YELLOW}Next step: Run AI subagent to fill inventories and diagrams${NC}"
