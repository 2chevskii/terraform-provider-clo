#!/usr/bin/env bash
set -euo pipefail

target="$OCI_REPOSITORY:$PROVIDER_VERSION"
working_dir=$(mktemp -d "$RUNNER_TEMP/clo-release.XXXXXX")
trap 'rm -rf -- "$working_dir"' EXIT

if oras manifest fetch "$target" > "$working_dir/published.json" 2> "$working_dir/lookup.err"; then
  if jq -e --arg revision "$SOURCE_REVISION" '
    .artifactType == "application/vnd.opentofu.provider" and
    .annotations["org.opencontainers.image.revision"] == $revision
  ' "$working_dir/published.json" > /dev/null; then
    printf '%s is already published for %s\n' "$target" "$SOURCE_REVISION"
    exit 0
  fi
  printf '%s already exists with different release metadata\n' "$target" >&2
  exit 1
fi
if ! grep -qiE 'not found|MANIFEST_UNKNOWN|NAME_UNKNOWN|404 Not Found' "$working_dir/lookup.err"; then
  cat "$working_dir/lookup.err" >&2
  exit 1
fi

source_url="$GITHUB_SERVER_URL/$GITHUB_REPOSITORY"
source_dir=$PWD
layout="$working_dir/layout"
platforms=(linux_amd64 linux_arm64 darwin_amd64 darwin_arm64 windows_amd64 windows_arm64)
for platform in "${platforms[@]}"; do
  os=${platform%_*}
  arch=${platform#*_}
  binary="terraform-provider-clo_v$PROVIDER_VERSION"
  if [[ $os == windows ]]; then
    binary+=.exe
  fi
  archive="terraform-provider-clo_${PROVIDER_VERSION}_${platform}.zip"
  printf 'Building %s\n' "$platform"
  GOOS="$os" GOARCH="$arch" CGO_ENABLED=0 go build -mod=readonly -trimpath \
    -ldflags='-s -w' -o "$working_dir/$binary" "$source_dir"
  (
    cd "$working_dir"
    zip -q "$archive" "$binary"
    oras push --oci-layout "$layout:$platform" \
      --artifact-type application/vnd.opentofu.provider-target \
      --artifact-platform "$os/$arch" \
      --annotation "org.opencontainers.image.source=$source_url" \
      "$archive:archive/zip"
    rm -- "$binary" "$archive"
  )
done

oras manifest index create --oci-layout "$layout:$PROVIDER_VERSION" \
  --artifact-type application/vnd.opentofu.provider \
  --annotation "org.opencontainers.image.source=$source_url" \
  --annotation "org.opencontainers.image.revision=$SOURCE_REVISION" \
  --annotation "org.opencontainers.image.version=$PROVIDER_VERSION" \
  "${platforms[@]}"
oras cp --from-oci-layout "$layout:$PROVIDER_VERSION" "$target"

expected=$(oras manifest fetch --oci-layout --descriptor "$layout:$PROVIDER_VERSION" | jq -er .digest)
published=$(oras manifest fetch --descriptor "$target" | jq -er .digest)
if [[ $published != "$expected" ]]; then
  printf 'Expected %s at %s, got %s\n' "$expected" "$target" "$published" >&2
  exit 1
fi
printf 'Published %s@%s\n' "$target" "$published"
