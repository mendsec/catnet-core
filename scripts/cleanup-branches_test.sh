#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/cleanup-branches.sh"

FAILED=0

assert_primary_protected() {
    local branch="$1"
    if is_primary_branch "$branch"; then
        echo "PASS: primary branch '$branch' is recognized as protected"
    else
        echo "FAIL: primary branch '$branch' should be recognized as protected"
        FAILED=1
    fi
}

assert_valid_branch() {
    local branch="$1"
    if is_valid_branch "$branch"; then
        echo "PASS: branch '$branch' is recognized as valid"
    else
        echo "FAIL: branch '$branch' should be valid"
        FAILED=1
    fi
}

assert_invalid_branch() {
    local branch="$1"
    if ! is_valid_branch "$branch"; then
        echo "PASS: branch '$branch' is recognized as invalid (eligible for cleanup)"
    else
        echo "FAIL: branch '$branch' should be invalid (eligible for cleanup)"
        FAILED=1
    fi
}

echo "=== Running cleanup-branches safety tests ==="

# 1. Verify primary branches are ALWAYS protected
assert_primary_protected "main"
assert_primary_protected "master"
assert_primary_protected "develop"

# 2. Verify valid DevSecOps branches
assert_valid_branch "main"
assert_valid_branch "develop"
assert_valid_branch "feature/scan-engine"
assert_valid_branch "bugfix/issue-123"
assert_valid_branch "hotfix/cve-fix"
assert_valid_branch "release/v0.4.0"
assert_valid_branch "security/audit-fix"
assert_valid_branch "dependabot/go_modules/golang.org/x/sys-0.45.0"

# 3. Verify non-conforming branches are marked invalid
assert_invalid_branch "temp-branch"
assert_invalid_branch "test-123"
assert_invalid_branch "random/branch"
assert_invalid_branch "develop-signed"
assert_invalid_branch "junk"

# 4. Verify dry-run execution on a mix of branches
echo "--- Testing dry-run cleanup output ---"
DRY_RUN=true cleanup_branches "main develop feature/foo random-junk"

if [ "$FAILED" -ne 0 ]; then
    echo "=== Some tests failed! ==="
    exit 1
fi

echo "=== All cleanup-branches tests passed successfully! ==="
