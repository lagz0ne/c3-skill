#!/bin/bash
# Build Table of Contents from C3 documentation
# Usage: .c3/scripts/build-toc.sh

set -e

C3_ROOT=".c3"
TOC_FILE="${C3_ROOT}/TOC.md"
TEMP_FILE=$(mktemp)

echo "Building C3 Table of Contents..."

# Header
cat > "$TEMP_FILE" << 'EOF'
# C3 Documentation Table of Contents

> **AUTO-GENERATED** - Do not edit manually. Regenerate with: `.c3/scripts/build-toc.sh`
>
EOF

echo "> Last generated: $(date '+%Y-%m-%d %H:%M:%S')" >> "$TEMP_FILE"
echo "" >> "$TEMP_FILE"

# Extract frontmatter field
extract_frontmatter() {
    local file=$1
    local field=$2
    awk -v field="$field" '
        /^---$/ { in_fm = !in_fm; next }
        in_fm && $0 ~ "^" field ":" {
            sub("^" field ": *", "")
            print
            exit
        }
    ' "$file"
}

# Extract multi-line summary
extract_summary() {
    local file=$1
    awk '
        /^---$/ { in_fm = !in_fm; next }
        in_fm && /^summary:/ {
            in_summary = 1
            sub(/^summary: *>? */, "")
            if (length($0) > 0) print
            next
        }
        in_fm && in_summary {
            if (/^[a-zA-Z-]+:/) { exit }
            if (length($0) > 0) print
        }
    ' "$file" | sed 's/^  //'
}

# Extract heading summary
extract_heading_summary() {
    local file=$1
    local heading_id=$2
    awk -v hid="$heading_id" '
        $0 ~ "{#" hid "}" {
            getline
            if ($0 ~ /^<!--/) {
                in_comment = 1
                sub(/^<!-- */, "")
                if ($0 !~ /-->$/) {
                    print
                    next
                }
            }
        }
        in_comment {
            if ($0 ~ /-->$/) {
                sub(/ *-->$/, "")
                if (length($0) > 0) print
                exit
            }
            print
        }
    ' "$file"
}

# List headings with IDs (POSIX-compliant, tab-delimited)
list_headings() {
    local file=$1
    awk '
        /^## .* \{#[a-z0-9-]+\}$/ {
            # Extract heading ID using POSIX-compliant gsub
            id_part = $0
            gsub(/.*\{#/, "", id_part)
            gsub(/\}.*/, "", id_part)
            heading_id = id_part

            # Extract heading title
            title_part = $0
            gsub(/ \{#[^}]+\}$/, "", title_part)
            gsub(/^## /, "", title_part)

            print heading_id "\t" title_part
        }
    ' "$file"
}

# Count documents
ctx_count=$(find "$C3_ROOT" -maxdepth 1 -name "CTX-*.md" 2>/dev/null | wc -l || echo 0)
con_count=$(find "$C3_ROOT/containers" -name "CON-*.md" 2>/dev/null | wc -l || echo 0)
com_count=$(find "$C3_ROOT/components" -name "COM-*.md" 2>/dev/null | wc -l || echo 0)
adr_count=$(find "$C3_ROOT/adr" -name "ADR-*.md" 2>/dev/null | wc -l || echo 0)

# Context Level
if [ "$ctx_count" -gt 0 ]; then
    echo "## Context Level" >> "$TEMP_FILE"
    echo "" >> "$TEMP_FILE"

    for file in $(find "$C3_ROOT" -maxdepth 1 -name "CTX-*.md" | sort); do
        id=$(extract_frontmatter "$file" "id")
        title=$(extract_frontmatter "$file" "title")
        summary=$(extract_summary "$file")

        echo "### [$id](./${id}.md) - $title" >> "$TEMP_FILE"
        echo "> $summary" >> "$TEMP_FILE"
        echo "" >> "$TEMP_FILE"

        echo "**Sections**:" >> "$TEMP_FILE"
        while IFS=$'\t' read -r heading_id heading_title; do
            heading_summary=$(extract_heading_summary "$file" "$heading_id")
            if [ -n "$heading_summary" ]; then
                echo "- [$heading_title](#$heading_id) - $heading_summary" >> "$TEMP_FILE"
            else
                echo "- [$heading_title](#$heading_id)" >> "$TEMP_FILE"
            fi
        done < <(list_headings "$file")
        echo "" >> "$TEMP_FILE"
        echo "---" >> "$TEMP_FILE"
        echo "" >> "$TEMP_FILE"
    done
fi

# Container Level
if [ "$con_count" -gt 0 ]; then
    echo "## Container Level" >> "$TEMP_FILE"
    echo "" >> "$TEMP_FILE"

    for file in $(find "$C3_ROOT/containers" -name "CON-*.md" 2>/dev/null | sort); do
        id=$(extract_frontmatter "$file" "id")
        title=$(extract_frontmatter "$file" "title")
        summary=$(extract_summary "$file")

        echo "### [$id](./containers/${id}.md) - $title" >> "$TEMP_FILE"
        echo "> $summary" >> "$TEMP_FILE"
        echo "" >> "$TEMP_FILE"

        echo "**Sections**:" >> "$TEMP_FILE"
        while IFS=$'\t' read -r heading_id heading_title; do
            heading_summary=$(extract_heading_summary "$file" "$heading_id")
            if [ -n "$heading_summary" ]; then
                echo "- [$heading_title](#$heading_id) - $heading_summary" >> "$TEMP_FILE"
            else
                echo "- [$heading_title](#$heading_id)" >> "$TEMP_FILE"
            fi
        done < <(list_headings "$file")
        echo "" >> "$TEMP_FILE"
        echo "---" >> "$TEMP_FILE"
        echo "" >> "$TEMP_FILE"
    done
fi

# Component Level
if [ "$com_count" -gt 0 ]; then
    echo "## Component Level" >> "$TEMP_FILE"
    echo "" >> "$TEMP_FILE"

    for container_dir in $(find "$C3_ROOT/components" -mindepth 1 -maxdepth 1 -type d 2>/dev/null | sort); do
        container_name=$(basename "$container_dir")

        echo "### ${container_name^} Components" >> "$TEMP_FILE"
        echo "" >> "$TEMP_FILE"

        for file in $(find "$container_dir" -name "COM-*.md" | sort); do
            id=$(extract_frontmatter "$file" "id")
            title=$(extract_frontmatter "$file" "title")
            summary=$(extract_summary "$file")

            echo "#### [$id](./components/$container_name/${id}.md) - $title" >> "$TEMP_FILE"
            echo "> $summary" >> "$TEMP_FILE"
            echo "" >> "$TEMP_FILE"

            echo "**Sections**:" >> "$TEMP_FILE"
            while IFS=$'\t' read -r heading_id heading_title; do
                heading_summary=$(extract_heading_summary "$file" "$heading_id")
                if [ -n "$heading_summary" ]; then
                    echo "- [$heading_title](#$heading_id) - $heading_summary" >> "$TEMP_FILE"
                else
                    echo "- [$heading_title](#$heading_id)" >> "$TEMP_FILE"
                fi
            done < <(list_headings "$file")
            echo "" >> "$TEMP_FILE"
            echo "---" >> "$TEMP_FILE"
            echo "" >> "$TEMP_FILE"
        done
    done
fi

# ADR Level
if [ "$adr_count" -gt 0 ]; then
    echo "## Architecture Decisions" >> "$TEMP_FILE"
    echo "" >> "$TEMP_FILE"

    for file in $(find "$C3_ROOT/adr" -name "ADR-*.md" 2>/dev/null | sort -r); do
        id=$(extract_frontmatter "$file" "id")
        title=$(extract_frontmatter "$file" "title")
        summary=$(extract_summary "$file")
        status=$(extract_frontmatter "$file" "status")

        echo "### [$id](./adr/${id}.md) - $title" >> "$TEMP_FILE"
        echo "> $summary" >> "$TEMP_FILE"
        echo "" >> "$TEMP_FILE"
        echo "**Status**: ${status^}" >> "$TEMP_FILE"
        echo "" >> "$TEMP_FILE"

        echo "**Sections**:" >> "$TEMP_FILE"
        while IFS=$'\t' read -r heading_id heading_title; do
            heading_summary=$(extract_heading_summary "$file" "$heading_id")
            if [ -n "$heading_summary" ]; then
                echo "- [$heading_title](#$heading_id) - $heading_summary" >> "$TEMP_FILE"
            else
                echo "- [$heading_title](#$heading_id)" >> "$TEMP_FILE"
            fi
        done < <(list_headings "$file")
        echo "" >> "$TEMP_FILE"
        echo "---" >> "$TEMP_FILE"
        echo "" >> "$TEMP_FILE"
    done
fi

# Quick Reference
total_docs=$((ctx_count + con_count + com_count + adr_count))
echo "## Quick Reference" >> "$TEMP_FILE"
echo "" >> "$TEMP_FILE"
echo "**Total Documents**: $total_docs" >> "$TEMP_FILE"
echo "**Contexts**: $ctx_count | **Containers**: $con_count | **Components**: $com_count | **ADRs**: $adr_count" >> "$TEMP_FILE"

# Move to final location
mv "$TEMP_FILE" "$TOC_FILE"

echo "Done. TOC generated: $TOC_FILE"
