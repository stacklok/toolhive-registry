#!/usr/bin/env bash
# Audit GitHub Actions workflows for security issues
# Based on Clinejection + hackerbot-claw campaign patterns
# Usage: ./audit-workflows.sh [workflow-dir]

set -euo pipefail

WORKFLOW_DIR="${1:-.github/workflows}"
ISSUES_FOUND=0
CRIT_COUNT=0
WARN_COUNT=0
RED='\033[0;31m'
YELLOW='\033[0;33m'
GREEN='\033[0;32m'
CYAN='\033[0;36m'
NC='\033[0m'

warn() { echo -e "${YELLOW}[WARN]${NC} $1"; ((WARN_COUNT++)); ((ISSUES_FOUND++)); }
crit() { echo -e "${RED}[CRIT]${NC} $1"; ((CRIT_COUNT++)); ((ISSUES_FOUND++)); }
ok()   { echo -e "${GREEN}[ OK ]${NC} $1"; }
info() { echo -e "${CYAN}[INFO]${NC} $1"; }

if [ ! -d "$WORKFLOW_DIR" ]; then
  echo "No workflow directory found at $WORKFLOW_DIR"
  exit 0
fi

echo "============================================"
echo "  CI/CD Security Audit"
echo "  Based on Clinejection + hackerbot-claw"
echo "============================================"
echo "Scanning: $WORKFLOW_DIR"
echo ""

# ============================================================
# Check 1: pull_request_target + fork code checkout (PWN REQUEST)
# This is the #1 attack vector — 4 of 7 hackerbot-claw attacks
# ============================================================
echo "--- [1/8] Pwn Request: pull_request_target + Fork Checkout ---"
for f in "$WORKFLOW_DIR"/*.yml "$WORKFLOW_DIR"/*.yaml; do
  [ -f "$f" ] || continue
  basename="$(basename "$f")"

  if grep -qE 'pull_request_target' "$f" 2>/dev/null; then
    # Check if it checks out fork code (PR head ref/sha)
    if grep -E '(pull_request\.head\.sha|pull_request\.head\.ref|github\.head_ref)' "$f" >/dev/null 2>&1; then
      crit "$basename: pull_request_target + fork code checkout = PWN REQUEST (Trivy/awesome-go attack vector)"
    else
      warn "$basename: Uses pull_request_target — verify it does NOT checkout fork code"
    fi
  fi
done
echo ""

# ============================================================
# Check 2: Expression injection in run: blocks
# Branch names, filenames, PR titles can contain shell payloads
# ============================================================
echo "--- [2/8] Expression Injection in run: Blocks ---"
for f in "$WORKFLOW_DIR"/*.yml "$WORKFLOW_DIR"/*.yaml; do
  [ -f "$f" ] || continue
  basename="$(basename "$f")"

  # Look for ${{ }} expressions referencing user-controlled fields inside run: blocks
  # We check for these fields anywhere in the file as a heuristic — the dangerous case
  # is when they appear in run: blocks, but that requires multi-line parsing
  if grep -E '\$\{\{\s*github\.event\.(issue|pull_request|comment|discussion|review|head_commit)\.(title|body|message|head\.ref)' "$f" >/dev/null 2>&1; then
    # Check if these appear near run: blocks (rough heuristic)
    if grep -B5 -A5 'run:' "$f" 2>/dev/null | grep -qE '\$\{\{\s*github\.event\.(issue|pull_request|comment|discussion|review|head_commit)\.(title|body|message|head\.ref)'; then
      crit "$basename: User-controlled input in run: block — shell injection (branch name / PR title attack vector)"
    else
      warn "$basename: References user-controlled fields via \${{ }} — verify they are NOT in run: blocks"
    fi
  fi

  # Check for head.ref / head_ref in run blocks (branch name injection)
  if grep -B5 -A5 'run:' "$f" 2>/dev/null | grep -qE '\$\{\{\s*(github\.head_ref|github\.event\.pull_request\.head\.ref)'; then
    crit "$basename: Branch name interpolated in run: block — branch name injection (microsoft/ai-discovery-agent attack vector)"
  fi
done
echo ""

# ============================================================
# Check 3: Slash commands without authorization
# /version, /deploy, /sync-metadata exploited in akri + DataDog
# ============================================================
echo "--- [3/8] Slash Command Authorization ---"
for f in "$WORKFLOW_DIR"/*.yml "$WORKFLOW_DIR"/*.yaml; do
  [ -f "$f" ] || continue
  basename="$(basename "$f")"

  if grep -qE 'issue_comment' "$f" 2>/dev/null; then
    if grep -qE "contains.*comment\.body.*'/'" "$f" 2>/dev/null || grep -qE "startsWith.*comment\.body.*'/'" "$f" 2>/dev/null; then
      if grep -qE 'author_association' "$f" 2>/dev/null; then
        ok "$basename: Slash command has author_association check"
      else
        crit "$basename: Slash command without author_association check — any user can trigger (akri/DataDog attack vector)"
      fi
    fi
  fi
done
echo ""

# ============================================================
# Check 4: AI agent configuration
# Prompt injection, tool access, user scoping
# ============================================================
echo "--- [4/8] AI Agent Configuration ---"
for f in "$WORKFLOW_DIR"/*.yml "$WORKFLOW_DIR"/*.yaml; do
  [ -f "$f" ] || continue
  basename="$(basename "$f")"

  if grep -qiE '(claude|copilot|openai|anthropic|ai-action|llm|code-action)' "$f" 2>/dev/null; then
    info "$basename: Contains AI agent integration"

    # Check for user-controlled input in prompts
    if grep -E '\$\{\{\s*github\.event\.(issue|pull_request|comment|discussion|review|head_commit)\.(title|body|message|ref)' "$f" >/dev/null 2>&1; then
      crit "$basename: User-controlled input interpolated into AI agent prompt (Clinejection attack vector)"
    fi

    # Check for overly permissive user access
    if grep -q 'allowed_non_write_users.*\*' "$f" 2>/dev/null; then
      crit "$basename: allowed_non_write_users is '*' — any user can trigger AI agent"
    fi

    # Check for dangerous tool grants
    if grep -qiE 'allowed_tools.*Bash|tools.*Bash' "$f" 2>/dev/null; then
      crit "$basename: AI agent has Bash access — enables arbitrary code execution"
    fi
    if grep -qiE 'allowed_tools.*(Write|Edit)|tools.*(Write|Edit)' "$f" 2>/dev/null; then
      warn "$basename: AI agent has Write/Edit access — can modify repository files"
    fi

    # Check if pull_request_target is used with AI (ambient-code attack vector)
    if grep -qE 'pull_request_target' "$f" 2>/dev/null; then
      warn "$basename: AI agent on pull_request_target — fork code may poison CLAUDE.md context"
    fi
  fi
done
echo ""

# ============================================================
# Check 5: Workflow permissions
# Missing or overly broad permissions
# ============================================================
echo "--- [5/8] Workflow Permissions ---"
for f in "$WORKFLOW_DIR"/*.yml "$WORKFLOW_DIR"/*.yaml; do
  [ -f "$f" ] || continue
  basename="$(basename "$f")"

  if grep -qE 'permissions:' "$f" 2>/dev/null; then
    if grep -qE 'write-all|permissions:\s*write' "$f" 2>/dev/null; then
      warn "$basename: Uses write-all permissions — apply least privilege"
    elif grep -qE 'contents:\s*write' "$f" 2>/dev/null; then
      # Check if it actually needs write (publish/release workflows)
      if ! grep -qiE '(publish|release|deploy|push|merge)' "$f" 2>/dev/null; then
        warn "$basename: Has contents: write but doesn't appear to publish/release — may be over-privileged"
      fi
    fi
  else
    warn "$basename: No explicit permissions: block — using repo defaults (may be too broad)"
  fi
done
echo ""

# ============================================================
# Check 6: Cache usage in release/publish workflows
# Cache poisoning via Cacheract (Clinejection)
# ============================================================
echo "--- [6/8] Cache Security ---"
for f in "$WORKFLOW_DIR"/*.yml "$WORKFLOW_DIR"/*.yaml; do
  [ -f "$f" ] || continue
  basename="$(basename "$f")"

  if grep -qiE '(npm publish|npx.*publish|vsce publish|ovsx publish|release|deploy|pypi|twine|goreleaser)' "$f" 2>/dev/null; then
    if grep -q 'actions/cache' "$f" 2>/dev/null; then
      warn "$basename: Release/publish workflow uses actions/cache — vulnerable to cache poisoning"
    else
      ok "$basename: Release workflow does not use actions/cache"
    fi
  fi
done
echo ""

# ============================================================
# Check 7: Publishing credential hygiene
# Provenance, token separation, multiple tokens
# ============================================================
echo "--- [7/8] Publishing & Credential Hygiene ---"
for f in "$WORKFLOW_DIR"/*.yml "$WORKFLOW_DIR"/*.yaml; do
  [ -f "$f" ] || continue
  basename="$(basename "$f")"

  # npm provenance
  if grep -qiE 'npm publish' "$f" 2>/dev/null; then
    if grep -q '\-\-provenance' "$f" 2>/dev/null; then
      ok "$basename: npm publish uses --provenance flag"
    else
      warn "$basename: npm publish without --provenance — no OIDC attestation"
    fi
  fi

  # Multiple high-privilege tokens in same workflow
  token_count=$(grep -coE 'secrets\.(NPM|VSCE|OVSX|PYPI|RUBYGEMS|DOCKER|PAT|TOKEN)' "$f" 2>/dev/null || echo 0)
  if [ "$token_count" -gt 2 ]; then
    warn "$basename: $token_count secret references in one workflow — consider separating into dedicated workflows"
  fi

  # PAT usage (stolen PAT = full compromise, as in Trivy)
  if grep -qiE 'secrets\.\w*PAT\w*' "$f" 2>/dev/null; then
    warn "$basename: Uses a PAT secret — if stolen, attacker gets full access (Trivy attack vector). Prefer GITHUB_TOKEN or fine-grained PATs"
  fi
done
echo ""

# ============================================================
# Check 8: Repository security posture
# SECURITY.md, CODEOWNERS, package hooks
# ============================================================
echo "--- [8/8] Repository Security Posture ---"

# SECURITY.md
if [ -f "SECURITY.md" ] || [ -f ".github/SECURITY.md" ]; then
  ok "SECURITY.md exists"
else
  warn "No SECURITY.md found — add vulnerability disclosure instructions"
fi

# CODEOWNERS for sensitive files
if [ -f "CODEOWNERS" ] || [ -f ".github/CODEOWNERS" ] || [ -f "docs/CODEOWNERS" ]; then
  codeowners_file=$(find . -maxdepth 2 -name CODEOWNERS 2>/dev/null | head -1)
  if [ -n "$codeowners_file" ]; then
    if grep -qiE '(CLAUDE\.md|\.claude|\.github/workflows)' "$codeowners_file" 2>/dev/null; then
      ok "CODEOWNERS protects sensitive files (CLAUDE.md / workflows)"
    else
      warn "CODEOWNERS exists but does not protect CLAUDE.md or .github/workflows/"
    fi
  fi
else
  warn "No CODEOWNERS file — add protection for CLAUDE.md and .github/workflows/"
fi

# Package.json hooks
if [ -f "package.json" ]; then
  for hook in preinstall install postinstall preuninstall postuninstall; do
    if grep -q "\"$hook\"" package.json 2>/dev/null; then
      warn "package.json: '$hook' script detected — review for unexpected commands"
    fi
  done
fi

# CLAUDE.md exists and could be targeted
if [ -f "CLAUDE.md" ]; then
  info "CLAUDE.md exists — ensure it is protected in CODEOWNERS (ambient-code attack vector)"
fi
echo ""

# ============================================================
# Summary
# ============================================================
echo "============================================"
echo "  Audit Complete"
echo "============================================"
if [ "$ISSUES_FOUND" -eq 0 ]; then
  echo -e "${GREEN}No issues found.${NC}"
else
  echo -e "${RED}Critical: $CRIT_COUNT${NC}  ${YELLOW}Warning: $WARN_COUNT${NC}  Total: $ISSUES_FOUND"
  if [ "$CRIT_COUNT" -gt 0 ]; then
    echo ""
    echo "CRITICAL findings are actively exploitable. Fix these first."
    echo "See references/ATTACK-PATTERNS.md for exploit details."
  fi
fi
exit "$ISSUES_FOUND"
