#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
DOC_DIR="${ROOT_DIR}/docs/api"
CURRENT_DOC="${DOC_DIR}/current.md"
CHANGELOG="${DOC_DIR}/changelog.md"
ENDPOINTS="${DOC_DIR}/endpoints.yaml"
STYLE_GUIDE="${DOC_DIR}/STYLE_GUIDE.md"

for file in "${CURRENT_DOC}" "${CHANGELOG}" "${ENDPOINTS}" "${STYLE_GUIDE}"; do
  if [[ ! -s "${file}" ]]; then
    echo "missing required docs file: ${file}" >&2
    exit 1
  fi
done

forbidden_pattern="$(printf '\u5ba2\u6237|\u7528\u6237')"
if rg -n "${forbidden_pattern}" "${CURRENT_DOC}" "${CHANGELOG}" "${STYLE_GUIDE}" >/tmp/api-docs-forbidden.txt; then
  cat /tmp/api-docs-forbidden.txt >&2
  echo "docs wording check failed" >&2
  exit 1
fi

for marker in "完整请求示例" "成功响应示例" "失败响应示例" "兼容性说明"; do
  if ! rg -q "${marker}" "${CURRENT_DOC}"; then
    echo "missing required section marker: ${marker}" >&2
    exit 1
  fi
done

version="$(sed -n 's/^version:[[:space:]]*"\{0,1\}\([^"]*\)"\{0,1\}$/\1/p' "${ENDPOINTS}" | head -1)"
if [[ -z "${version}" ]]; then
  echo "missing version in endpoints.yaml" >&2
  exit 1
fi

if ! rg -q "${version}" "${CHANGELOG}"; then
  echo "changelog does not mention endpoint manifest version: ${version}" >&2
  exit 1
fi

while IFS= read -r path; do
  [[ -z "${path}" ]] && continue
  if ! rg -F -q "${path}" "${CURRENT_DOC}"; then
    echo "endpoint is missing from current docs: ${path}" >&2
    exit 1
  fi
done < <(sed -n 's/^[[:space:]]*path:[[:space:]]*//p' "${ENDPOINTS}")

echo "api docs check passed"
