#!/usr/bin/env bash
set -euo pipefail

# Allowed DevSecOps branch patterns
PATTERN="^(main|develop|feature/.*|bugfix/.*|hotfix/.*|release/.*|security/.*|dependabot/.*)$"
DRY_RUN="${DRY_RUN:-false}"

# Primary protected branches that must never be deleted under any circumstance
is_primary_branch() {
    local branch="$1"
    case "$branch" in
        main|master|develop)
            return 0
            ;;
        *)
            return 1
            ;;
    esac
}

is_valid_branch() {
    local branch="$1"
    if is_primary_branch "$branch"; then
        return 0
    fi
    if [[ "$branch" =~ $PATTERN ]]; then
        return 0
    fi
    return 1
}

cleanup_branches() {
    local branches="$1"
    for branch in $branches; do
        branch=$(echo "$branch" | xargs)
        if [ -z "$branch" ]; then
            continue
        fi

        if is_primary_branch "$branch"; then
            echo "Branch '$branch' is a primary protected branch. Skipping."
            continue
        fi

        if [[ ! "$branch" =~ $PATTERN ]]; then
            echo "Branch '$branch' does not match DevSecOps culture. Deleting..."
            if [ "$DRY_RUN" = "true" ]; then
                echo "[DRY-RUN] Would delete remote branch '$branch'"
            else
                git push origin --delete "$branch" || echo "Failed to delete $branch (it may be protected)"
            fi
        else
            echo "Branch '$branch' is valid."
        fi
    done
}

if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    # Fetch all remote branches (excluding HEAD pointer)
    REMOTE_BRANCHES=$(git branch -r | grep -v '\->' | sed 's/origin\///')
    cleanup_branches "$REMOTE_BRANCHES"
fi
