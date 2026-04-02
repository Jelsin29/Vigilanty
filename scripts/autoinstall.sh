#!/usr/bin/env bash

set -euo pipefail

REPO="Jelsin29/Vigilanty"
APP_NAME="vigilanty"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
REQUESTED_VERSION="${VIGILANTY_VERSION:-}"
TMP_DIR=""

log() {
	printf '[vigilanty-install] %s\n' "$*"
}

fail() {
	printf '[vigilanty-install] ERROR: %s\n' "$*" >&2
	exit 1
}

need_cmd() {
	command -v "$1" >/dev/null 2>&1 || fail "Missing required command: $1"
}

normalize_os() {
	local os
	os="$(uname -s)"
	case "$os" in
		Linux) echo "linux" ;;
		Darwin) echo "darwin" ;;
		*) fail "Unsupported OS: $os (supported: Linux, Darwin)" ;;
	esac
}

normalize_arch() {
	local arch
	arch="$(uname -m)"
	case "$arch" in
		x86_64|amd64) echo "amd64" ;;
		arm64|aarch64) echo "arm64" ;;
		*) fail "Unsupported architecture: $arch (supported: amd64, arm64)" ;;
	esac
}

pick_release_api() {
	local version="${1:-}"

	if [[ -z "$version" ]]; then
		echo "https://api.github.com/repos/${REPO}/releases/latest"
		return
	fi

	if [[ "$version" == v* ]]; then
		echo "https://api.github.com/repos/${REPO}/releases/tags/${version}"
	else
		echo "https://api.github.com/repos/${REPO}/releases/tags/v${version}"
	fi
}

extract_json_string() {
	local key="$1"
	local json="$2"
	printf '%s' "$json" | sed -n "s/.*\"${key}\"[[:space:]]*:[[:space:]]*\"\([^\"]*\)\".*/\1/p" | head -n1
}

select_asset_url() {
	local json="$1"
	local os="$2"
	local arch="$3"
	local arch_alt=""
	local url

	case "$arch" in
		amd64) arch_alt="x86_64" ;;
		arm64) arch_alt="aarch64" ;;
	esac

	while IFS= read -r url; do
		local lower
		local filename
		lower="$(printf '%s' "$url" | tr '[:upper:]' '[:lower:]')"
		filename="${lower##*/}"

		[[ "$filename" == ${APP_NAME}* ]] || continue
		[[ "$filename" == *"-${os}-"* || "$filename" == *"_${os}_"* ]] || continue

		if [[ "$filename" != *"-${arch}."* && "$filename" != *"_${arch}."* && "$filename" != *"-${arch_alt}."* && "$filename" != *"_${arch_alt}."* ]]; then
			continue
		fi

		if [[ "$filename" == *.tar.gz || "$filename" == *.tgz || "$filename" == *.zip ]]; then
			printf '%s\n' "$url"
			return 0
		fi
	done < <(printf '%s' "$json" | sed -n 's/.*"browser_download_url"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p')

	return 1
}

install_binary() {
	local bin_path="$1"
	local target_path="${INSTALL_DIR}/${APP_NAME}"

	chmod +x "$bin_path"

	if [[ -w "$INSTALL_DIR" ]]; then
		install -m 0755 "$bin_path" "$target_path"
	else
		need_cmd sudo
		sudo install -m 0755 "$bin_path" "$target_path"
	fi

	log "Installed to ${target_path}"
}

main() {
	need_cmd curl
	need_cmd tar
	need_cmd uname
	need_cmd sed
	need_cmd install

	local os arch api_url release_json tag_name asset_url archive_path extract_dir bin_path
	os="$(normalize_os)"
	arch="$(normalize_arch)"
	api_url="$(pick_release_api "$REQUESTED_VERSION")"

	log "Detecting platform: ${os}/${arch}"
	log "Checking release metadata from GitHub"
	release_json="$(curl -fsSL "$api_url")" || fail "Unable to fetch release metadata from ${api_url}"

	tag_name="$(extract_json_string tag_name "$release_json")"
	[[ -n "$tag_name" ]] || fail "Release metadata missing tag_name. Check release/tag exists and repo visibility."

	asset_url="$(select_asset_url "$release_json" "$os" "$arch")" || fail "No release asset found for ${os}/${arch}"

	TMP_DIR="$(mktemp -d)"
	trap 'if [[ -n "${TMP_DIR:-}" ]]; then rm -rf "$TMP_DIR"; fi' EXIT

	archive_path="${TMP_DIR}/archive"
	extract_dir="${TMP_DIR}/extract"
	mkdir -p "$extract_dir"

	log "Downloading ${tag_name} asset"
	curl -fsSL -o "$archive_path" "$asset_url"

	case "$asset_url" in
		*.tar.gz|*.tgz)
			tar -xzf "$archive_path" -C "$extract_dir"
			;;
		*.zip)
			need_cmd unzip
			unzip -q "$archive_path" -d "$extract_dir"
			;;
		*)
			fail "Unsupported asset format: $asset_url"
			;;
	esac

	bin_path="$(find "$extract_dir" -type f \( -name "$APP_NAME" -o -name "${APP_NAME}-*" \) | head -n1)"
	[[ -n "$bin_path" ]] || fail "Binary '${APP_NAME}' not found in extracted archive"

	install_binary "$bin_path"

	log "Done. Run '${APP_NAME} version' to verify installation."
}

main "$@"
