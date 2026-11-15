#!/usr/bin/env bash
# Decide the default-checks job result from one go test -json run.
#
# Usage: default-checks-result.sh RESULTS STATUS
#
# A package with no test files emits "Action":"skip" and no "Test" field.
# That event does not fail the job. A skipped test emits both keys on one
# line. go test -json currently writes Time, Action, Package, then Test.
# The match below accepts either key order on that line, so a future encoder
# that reorders keys still counts a named skip. A non-zero STATUS still
# fails the job when no test was skipped.

set -eu

if [[ $# -ne 2 ]]; then
  echo "usage: default-checks-result.sh RESULTS STATUS" >&2
  exit 2
fi

results=$1
status=$2

if grep -E '"Action":"skip".*"Test":|"Test":.*"Action":"skip"' "$results" >/dev/null; then
  echo "A check was skipped"
  exit 1
fi

exit "$status"
