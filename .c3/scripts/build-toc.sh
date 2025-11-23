#!/bin/bash
# Build Table of Contents from C3 documentation
# Usage: .c3/scripts/build-toc.sh
#
# Supports:
# - v2: containers/ and components/ folders, VERSION file
# - v3: c3-[0-9]-*/ container folders, c3-version in README.md frontmatter

set -e

C3_ROOT=".c3"
TOC_FILE="${C3_ROOT}/TOC.md"
TEMP_FILE=$(mktemp)

echo "Building C3 Table of Contents..."

# Detect version - v3 uses container folders, v2 uses containers/ directory
detect_version() {
    # Check for v3 container folders (c3-[0-9]-*/)
    v3_container_count=$(find "$C3_ROOT" -maxdepth 1 -type d -name "c3-[0-9]-*" 2>/dev/null | wc -l || echo 0)

    # Check for v2 containers directory
    v2_containers_exist=0
    [ -d "$C3_ROOT/containers" ] && v2_containers_exist=1

    if [ "$v3_container_count" -gt 0 ]; then
        echo "v3"
    elif [ "$v2_containers_exist" -eq 1 ]; then
        echo "v2"
    else
        # Default to v3 if no containers found (new project)
        echo "v3"
    fi
}

C3_VERSION=$(detect_version)
echo "Detected C3 structure: $C3_VERSION"

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

# Output sections for a file
output_sections() {
    local file=$1
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
}

# Count documents based on version
if [ "$C3_VERSION" = "v3" ]; then
    # v3 counts
    readme_exists=0
    [ -f "$C3_ROOT/README.md" ] && readme_exists=1
    ctx_count=$(find "$C3_ROOT" -maxdepth 1 -name "ctx-*.md" 2>/dev/null | wc -l || echo 0)
    ctx_total=$((readme_exists + ctx_count))

    con_count=$(find "$C3_ROOT" -maxdepth 1 -type d -name "c3-[0-9]-*" 2>/dev/null | wc -l || echo 0)
    com_count=$(find "$C3_ROOT" -path "*/c3-[0-9]-*/c3-[0-9][0-9][0-9]-*.md" 2>/dev/null | wc -l || echo 0)
    adr_count=$(find "$C3_ROOT/adr" -name "adr-*.md" 2>/dev/null | wc -l || echo 0)
else
    # v2 counts (existing logic)
    readme_exists=0
    [ -f "$C3_ROOT/README.md" ] && readme_exists=1
    ctx_count=$(find "$C3_ROOT" -maxdepth 1 -name "CTX-*.md" 2>/dev/null | wc -l || echo 0)
    ctx_total=$((readme_exists + ctx_count))

    con_count=$(find "$C3_ROOT/containers" -name "C3-[0-9]-*.md" 2>/dev/null | wc -l || echo 0)

    # Component count - check flat first, then nested
    flat_com_count=$(find "$C3_ROOT/components" -maxdepth 1 -name "C3-[0-9][0-9][0-9]-*.md" 2>/dev/null | wc -l || echo 0)
    nested_com_count=$(find "$C3_ROOT/components" -mindepth 2 -name "C3-[0-9][0-9][0-9]-*.md" 2>/dev/null | wc -l || echo 0)
    com_count=$((flat_com_count + nested_com_count))

    adr_count=$(find "$C3_ROOT/adr" -name "ADR-*.md" 2>/dev/null | wc -l || echo 0)
fi

# Context Level - check README.md first, then auxiliary context files
readme_file="$C3_ROOT/README.md"
readme_exists=0
[ -f "$readme_file" ] && readme_exists=1

if [ "$C3_VERSION" = "v3" ]; then
    ctx_files=$(find "$C3_ROOT" -maxdepth 1 -name "ctx-*.md" 2>/dev/null | sort)
else
    ctx_files=$(find "$C3_ROOT" -maxdepth 1 -name "CTX-*.md" 2>/dev/null | sort)
fi

if [ "$readme_exists" -eq 1 ] || [ -n "$ctx_files" ]; then
    echo "## Context Level" >> "$TEMP_FILE"
    echo "" >> "$TEMP_FILE"

    # Process README.md as primary context
    if [ "$readme_exists" -eq 1 ]; then
        id=$(extract_frontmatter "$readme_file" "id")
        title=$(extract_frontmatter "$readme_file" "title")
        summary=$(extract_summary "$readme_file")

        echo "### [$id](./README.md) - $title" >> "$TEMP_FILE"
        echo "> $summary" >> "$TEMP_FILE"
        echo "" >> "$TEMP_FILE"

        output_sections "$readme_file"
    fi

    # Process auxiliary context files
    for file in $ctx_files; do
        id=$(extract_frontmatter "$file" "id")
        title=$(extract_frontmatter "$file" "title")
        summary=$(extract_summary "$file")
        filename=$(basename "$file")

        echo "### [$id](./${filename}) - $title" >> "$TEMP_FILE"
        echo "> $summary" >> "$TEMP_FILE"
        echo "" >> "$TEMP_FILE"

        output_sections "$file"
    done
fi

# Container Level
if [ "$C3_VERSION" = "v3" ]; then
    # v3: Container folders with README.md inside
    container_dirs=$(find "$C3_ROOT" -maxdepth 1 -type d -name "c3-[0-9]-*" 2>/dev/null | sort)

    if [ -n "$container_dirs" ]; then
        echo "## Container Level" >> "$TEMP_FILE"
        echo "" >> "$TEMP_FILE"

        for container_dir in $container_dirs; do
            container_name=$(basename "$container_dir")
            container_readme="$container_dir/README.md"

            if [ -f "$container_readme" ]; then
                id=$(extract_frontmatter "$container_readme" "id")
                title=$(extract_frontmatter "$container_readme" "title")
                summary=$(extract_summary "$container_readme")

                echo "### [$id](./${container_name}/) - $title" >> "$TEMP_FILE"
                echo "> $summary" >> "$TEMP_FILE"
                echo "" >> "$TEMP_FILE"

                output_sections "$container_readme"
            fi
        done
    fi
else
    # v2: containers/ directory with .md files
    if [ "$con_count" -gt 0 ]; then
        echo "## Container Level" >> "$TEMP_FILE"
        echo "" >> "$TEMP_FILE"

        for file in $(find "$C3_ROOT/containers" -name "C3-[0-9]-*.md" 2>/dev/null | sort); do
            id=$(extract_frontmatter "$file" "id")
            title=$(extract_frontmatter "$file" "title")
            summary=$(extract_summary "$file")

            echo "### [$id](./containers/${id}.md) - $title" >> "$TEMP_FILE"
            echo "> $summary" >> "$TEMP_FILE"
            echo "" >> "$TEMP_FILE"

            output_sections "$file"
        done
    fi
fi

# Component Level
if [ "$C3_VERSION" = "v3" ]; then
    # v3: Components inside container folders
    container_dirs=$(find "$C3_ROOT" -maxdepth 1 -type d -name "c3-[0-9]-*" 2>/dev/null | sort)
    total_components=0

    # Check if any components exist
    for container_dir in $container_dirs; do
        comp_count=$(find "$container_dir" -maxdepth 1 -name "c3-[0-9][0-9][0-9]-*.md" 2>/dev/null | wc -l || echo 0)
        total_components=$((total_components + comp_count))
    done

    if [ "$total_components" -gt 0 ]; then
        echo "## Component Level" >> "$TEMP_FILE"
        echo "" >> "$TEMP_FILE"

        for container_dir in $container_dirs; do
            container_name=$(basename "$container_dir")
            container_readme="$container_dir/README.md"

            # Get container title for grouping
            if [ -f "$container_readme" ]; then
                container_title=$(extract_frontmatter "$container_readme" "title")
            else
                container_title="$container_name"
            fi

            component_files=$(find "$container_dir" -maxdepth 1 -name "c3-[0-9][0-9][0-9]-*.md" 2>/dev/null | sort)

            if [ -n "$component_files" ]; then
                echo "### $container_title Components" >> "$TEMP_FILE"
                echo "" >> "$TEMP_FILE"

                for file in $component_files; do
                    id=$(extract_frontmatter "$file" "id")
                    title=$(extract_frontmatter "$file" "title")
                    summary=$(extract_summary "$file")
                    filename=$(basename "$file")

                    echo "#### [$id](./${container_name}/${filename}) - $title" >> "$TEMP_FILE"
                    echo "> $summary" >> "$TEMP_FILE"
                    echo "" >> "$TEMP_FILE"

                    output_sections "$file"
                done
            fi
        done
    fi
else
    # v2: Flat or nested components directory
    if [ "$flat_com_count" -gt 0 ]; then
        # V2: Flat structure - components directly in components/ directory
        echo "## Component Level" >> "$TEMP_FILE"
        echo "" >> "$TEMP_FILE"

        # Group by container number (first digit after C3-)
        for container_num in $(find "$C3_ROOT/components" -maxdepth 1 -name "C3-[0-9][0-9][0-9]-*.md" -exec basename {} \; 2>/dev/null | sed 's/C3-\([0-9]\).*/\1/' | sort -u); do
            echo "### Container $container_num Components" >> "$TEMP_FILE"
            echo "" >> "$TEMP_FILE"

            for file in $(find "$C3_ROOT/components" -maxdepth 1 -name "C3-${container_num}[0-9][0-9]-*.md" 2>/dev/null | sort); do
                id=$(extract_frontmatter "$file" "id")
                title=$(extract_frontmatter "$file" "title")
                summary=$(extract_summary "$file")

                echo "#### [$id](./components/${id}.md) - $title" >> "$TEMP_FILE"
                echo "> $summary" >> "$TEMP_FILE"
                echo "" >> "$TEMP_FILE"

                output_sections "$file"
            done
        done
    elif [ "$nested_com_count" -gt 0 ]; then
        # V1: Nested structure - components in container subfolders
        echo "## Component Level" >> "$TEMP_FILE"
        echo "" >> "$TEMP_FILE"

        for container_dir in $(find "$C3_ROOT/components" -mindepth 1 -maxdepth 1 -type d 2>/dev/null | sort); do
            container_name=$(basename "$container_dir")

            echo "### ${container_name^} Components" >> "$TEMP_FILE"
            echo "" >> "$TEMP_FILE"

            for file in $(find "$container_dir" -name "C3-[0-9][0-9][0-9]-*.md" | sort); do
                id=$(extract_frontmatter "$file" "id")
                title=$(extract_frontmatter "$file" "title")
                summary=$(extract_summary "$file")

                echo "#### [$id](./components/$container_name/${id}.md) - $title" >> "$TEMP_FILE"
                echo "> $summary" >> "$TEMP_FILE"
                echo "" >> "$TEMP_FILE"

                output_sections "$file"
            done
        done
    fi
fi

# ADR Level
if [ "$C3_VERSION" = "v3" ]; then
    # v3: lowercase adr-*.md
    if [ "$adr_count" -gt 0 ]; then
        echo "## Architecture Decisions" >> "$TEMP_FILE"
        echo "" >> "$TEMP_FILE"

        for file in $(find "$C3_ROOT/adr" -name "adr-*.md" 2>/dev/null | sort -r); do
            id=$(extract_frontmatter "$file" "id")
            title=$(extract_frontmatter "$file" "title")
            summary=$(extract_summary "$file")
            status=$(extract_frontmatter "$file" "status")
            filename=$(basename "$file")

            echo "### [$id](./adr/${filename}) - $title" >> "$TEMP_FILE"
            echo "> $summary" >> "$TEMP_FILE"
            echo "" >> "$TEMP_FILE"
            echo "**Status**: ${status^}" >> "$TEMP_FILE"
            echo "" >> "$TEMP_FILE"

            output_sections "$file"
        done
    fi
else
    # v2: uppercase ADR-*.md
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

            output_sections "$file"
        done
    fi
fi

# Quick Reference
total_docs=$((ctx_total + con_count + com_count + adr_count))
echo "## Quick Reference" >> "$TEMP_FILE"
echo "" >> "$TEMP_FILE"
echo "**Total Documents**: $total_docs" >> "$TEMP_FILE"
echo "**Contexts**: $ctx_total | **Containers**: $con_count | **Components**: $com_count | **ADRs**: $adr_count" >> "$TEMP_FILE"

# Move to final location
mv "$TEMP_FILE" "$TOC_FILE"

echo "Done. TOC generated: $TOC_FILE"
