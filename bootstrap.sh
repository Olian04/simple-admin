#!/usr/bin/env bash

set -euo pipefail

GUM_TOOL="github.com/charmbracelet/gum"
GUM=(go tool gum)
OLD_MODULE="github.com/Olian04/go-app-template"
OLD_CMD="echo"

remote_module_path() {
	local remote top
	top="$(git rev-parse --show-toplevel 2>/dev/null || true)"
	if [[ "${top}" != "$(pwd)" ]]; then
		return 1
	fi
	remote="$(git config --get remote.origin.url 2>/dev/null || true)"
	if [[ -z "${remote}" ]]; then
		return 1
	fi
	remote="${remote%.git}"
	case "${remote}" in
		git@*:*)
			remote="${remote#git@}"
			remote="${remote/:/\/}"
			;;
		ssh://git@*)
			remote="${remote#ssh://git@}"
			;;
		https://*)
			remote="${remote#https://}"
			;;
		http://*)
			remote="${remote#http://}"
			;;
	esac
	printf '%s\n' "${remote}"
}

module_basename() {
	local module_path="$1"
	local name="${module_path##*/}"
	name="${name%.git}"
	printf '%s\n' "${name,,}"
}

latest_go_major_minor() {
	local version
	version="${GO_VERSION:-$(go env GOVERSION)}"
	version="${version#go}"
	if [[ "${version}" =~ ^([0-9]+)\.([0-9]+) ]]; then
		printf '%s.%s\n' "${BASH_REMATCH[1]}" "${BASH_REMATCH[2]}"
		return 0
	fi
	printf 'unable to parse Go version from %q\n' "${version}" >&2
	return 1
}

skip_bootstrap_prompts() {
	# No TTY (piped/CI) or explicit opt-out: use MODULE_PATH / CMD_NAME / git inference only.
	if [[ "${BOOTSTRAP_NONINTERACTIVE:-0}" == "1" ]]; then
		return 0
	fi
	if [[ ! -t 0 ]]; then
		return 0
	fi
	return 1
}

require_clean_worktree() {
	if [[ "${BOOTSTRAP_ALLOW_DIRTY:-0}" == "1" ]]; then
		return 0
	fi
	if ! git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
		cat >&2 <<'EOF'
bootstrap: error: needs a git repository

This script rewrites the tree in place. Put the project in a git working tree
so the template snapshot is something you can go back to.

Initialize or clone a repository here, then run bootstrap again.
EOF
		exit 1
	fi
	if [[ -n "$(git status --porcelain 2>/dev/null)" ]]; then
		cat >&2 <<'EOF'
bootstrap: error: commit the initial template before bootstrapping

Bootstrap is destructive: it rewrites imports and paths, may rename or remove
cmd packages, and updates release-related files. If it stops halfway through,
you can be left in a broken, hard-to-fix state.

Commit the repository as-is then run this script again.
EOF
		exit 1
	fi
}

default_module_path_from_env_or_git() {
	if [[ -n "${MODULE_PATH:-}" ]]; then
		printf '%s\n' "${MODULE_PATH}"
		return 0
	fi
	remote_module_path
}

prompt_module_path() {
	local default_val="$1"
	local out
	out="$("${GUM[@]}" input \
		--header "Go module path" \
		--placeholder "e.g. github.com/you/your-app" \
		--value "${default_val}")"
	printf '%s\n' "${out}"
}

prompt_cmd_name() {
	local default_val="$1"
	local out
	out="$("${GUM[@]}" input \
		--header "Main command name (directory under cmd/, binary name in builds)" \
		--placeholder "e.g. myapp" \
		--value "${default_val}")"
	printf '%s\n' "${out}"
}

replace_all() {
	local from="$1"
	local to="$2"
	python3 - "$from" "$to" <<'PY'
import sys
from pathlib import Path

root = Path.cwd()
old, new = sys.argv[1], sys.argv[2]
skip_dirs = {".git", "dist"}
for path in root.rglob("*"):
    if any(part in skip_dirs for part in path.parts):
        continue
    if path.name == "bootstrap.sh":
        continue
    if not path.is_file():
        continue
    try:
        data = path.read_text()
    except UnicodeDecodeError:
        continue
    updated = data.replace(old, new)
    if updated != data:
        path.write_text(updated)
PY
}

# Rewrite Go import paths, -X ldflags, docs, etc.: cmd/<old>/... -> cmd/<new>/...
replace_cmd_package_paths() {
	local old_cmd="$1"
	local new_cmd="$2"
	if [[ "${old_cmd}" == "${new_cmd}" ]]; then
		return 0
	fi
	python3 - "$old_cmd" "$new_cmd" <<'PY'
import re
import sys
from pathlib import Path

old_cmd, new_cmd = sys.argv[1], sys.argv[2]
skip_dirs = {".git", "dist"}
pat = re.compile(r"cmd/" + re.escape(old_cmd) + r"(?![a-zA-Z0-9_-])")

for path in Path.cwd().rglob("*"):
    if any(part in skip_dirs for part in path.parts):
        continue
    if path.name == "bootstrap.sh":
        continue
    if not path.is_file():
        continue
    try:
        data = path.read_text()
    except UnicodeDecodeError:
        continue
    updated = pat.sub(f"cmd/{new_cmd}", data)
    if updated != data:
        path.write_text(updated)
PY
}

update_command_metadata() {
	local old="$1"
	local new="$2"
	python3 - "$old" "$new" <<'PY'
import re
import sys
from pathlib import Path

old, new = sys.argv[1], sys.argv[2]
files = [
    ".gitignore",
    "Makefile",
    "README.md",
    "docs/AGENT_CONTEXT.md",
    "goreleaser/.goreleaser.yaml",
    "goreleaser/Dockerfile",
]

replacements = {
    f"cmd/{old}": f"cmd/{new}",
    f"./dist/{old}": f"./dist/{new}",
    f"./{old}": f"./{new}",
    f"COPY {old} ": f"COPY {new} ",
    f"/usr/local/bin/{old}": f"/usr/local/bin/{new}",
    f"${{TARGETPLATFORM}}/{old}": f"${{TARGETPLATFORM}}/{new}",
    f"project_name: {old}-template": f"project_name: {new}",
    f"id: {old}": f"id: {new}",
    f"binary: {old}": f"binary: {new}",
}

for name in files:
    path = Path(name)
    if not path.exists():
        continue
    data = path.read_text(encoding="utf-8")
    updated = data
    for source, target in replacements.items():
        updated = updated.replace(source, target)
    if path.name == ".gitignore" and old != new:
        updated = updated.replace(f"/{old}", f"/{new}")
    if name == "goreleaser/.goreleaser.yaml":
        updated = re.sub(
            rf"^(\s*)- {re.escape(old)}\s*$",
            rf"\1- {new}",
            updated,
            flags=re.MULTILINE,
        )
    if updated != data:
        path.write_text(updated, encoding="utf-8")
PY
}

main() {
	local root module_path cmd_name go_version self
	local default_module_hint default_cmd_hint entered_module entered_cmd
	self="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/$(basename "${BASH_SOURCE[0]}")"
	root="$(dirname "${self}")"
	cd "${root}"

	require_clean_worktree

	default_module_hint=""
	if default_module_hint="$(default_module_path_from_env_or_git 2>/dev/null)"; then
		:
	fi

	if skip_bootstrap_prompts; then
		module_path="${MODULE_PATH:-${default_module_hint}}"
		if [[ -z "${module_path}" ]]; then
			"${GUM[@]}" log --level error "Could not determine module path" \
				"hint" "set MODULE_PATH or configure git remote, or run interactively in a terminal"
			exit 1
		fi
		cmd_name="${CMD_NAME:-$(module_basename "${module_path}")}"
	else
		entered_module="$(prompt_module_path "${default_module_hint}")"
		module_path="$(printf '%s\n' "${entered_module}" | sed -e 's/^[[:space:]]*//' -e 's/[[:space:]]*$//')"
		if [[ -z "${module_path}" ]]; then
			"${GUM[@]}" log --level error "Module path cannot be empty"
			exit 1
		fi
		default_cmd_hint="${CMD_NAME:-$(module_basename "${module_path}")}"
		entered_cmd="$(prompt_cmd_name "${default_cmd_hint}")"
		cmd_name="$(printf '%s\n' "${entered_cmd}" | sed -e 's/^[[:space:]]*//' -e 's/[[:space:]]*$//')"
		if [[ -z "${cmd_name}" ]]; then
			"${GUM[@]}" log --level error "Command name cannot be empty"
			exit 1
		fi
	fi

	go_version="$(latest_go_major_minor)"

	"${GUM[@]}" style --bold --border rounded --padding "1 2" "Bootstrapping ${module_path}"
	"${GUM[@]}" log --structured --level info "Resolved template settings" module "${module_path}" cmd "${cmd_name}" go "${go_version}"

	"${GUM[@]}" log --level info "Updating module path"
	replace_all "${OLD_MODULE}" "${module_path}"

	if [[ "${cmd_name}" != "${OLD_CMD}" && -d "cmd/${OLD_CMD}" ]]; then
		if [[ -d "cmd/${cmd_name}" ]]; then
			# mv cmd/old cmd/new when new already exists moves old *into* new (cmd/new/old); avoid that.
			if [[ "${BOOTSTRAP_SAFE:-0}" == "1" ]]; then
				"${GUM[@]}" log --level error "cmd/${cmd_name} already exists" \
					"hint" "delete cmd/${OLD_CMD} or move it aside, then re-run (or omit BOOTSTRAP_SAFE=1 to remove stale cmd/${OLD_CMD})"
				exit 1
			fi
			"${GUM[@]}" log --level warn "cmd/${cmd_name} already exists" "action" "removing stale cmd/${OLD_CMD}"
			rm -rf -- "cmd/${OLD_CMD}"
		else
			"${GUM[@]}" spin --spinner dot --show-error --title "Renaming command" -- mv "cmd/${OLD_CMD}" "cmd/${cmd_name}"
		fi
	fi

	if [[ "${cmd_name}" != "${OLD_CMD}" ]]; then
		"${GUM[@]}" log --level info "Updating cmd/ import and ldflag paths"
		replace_cmd_package_paths "${OLD_CMD}" "${cmd_name}"
	fi

	"${GUM[@]}" log --level info "Updating command references"
	if ! update_command_metadata "${OLD_CMD}" "${cmd_name}"; then
		"${GUM[@]}" log --level error "update_command_metadata failed" "hint" "see stdout/stderr above"
		exit 1
	fi
	"${GUM[@]}" spin --spinner dot --show-error --title "Updating Go version" -- go mod edit -go="${go_version}"
	"${GUM[@]}" spin --spinner dot --show-error --title "Tidying module" -- go mod tidy

	"${GUM[@]}" log --structured --level info "Template configured" module "${module_path}" cmd "${cmd_name}" go "${go_version}"
	if [[ "${BOOTSTRAP_KEEP:-0}" != "1" ]]; then
		"${GUM[@]}" log --level info "Removing bootstrap script" path "${self}"
		rm -- "${self}"
	fi

	# Bootstrap is done; gum is only used above. Drop it from `tool` and prune the module graph.
	printf '%s\n' "INFO Dropping ${GUM_TOOL} from module tools; running go mod tidy"
	go mod edit -droptool="${GUM_TOOL}"
	go mod tidy
}

main "$@"
