# Copyright 2020 The Knative Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

function build_flags() {
  local base="${1}"
  local now="$(date -u '+%Y-%m-%d %H:%M:%S')"
  local rev="$(git rev-parse --short HEAD)"
  local version="${TAG:-}"
  # Use vYYYYMMDD-local-<hash> for the version string, if not passed.
  if [[ -z "${version}" ]]; then
    # Get the commit, excluding any tags but keeping the "dirty" flag
    local commit="$(git describe --always --dirty --match '^$')"
    [[ -n "${commit}" ]] || abort "error getting the current commit"
    version="v$(date +%Y%m%d)-local-${commit}"
  fi

  local TMP_CORE=$(mktemp)
  local TMP_IMC=$(mktemp)
  # This is intentionally single-arch as it's for development.
  # release.yaml should embed the multi-arch version.
  ko resolve ${KOFLAGS:-} --tags ${version} -BRf config/core | ${PROCESSOR:-cat} > $TMP_CORE
  ko resolve ${KOFLAGS:-} --tags ${version} -BRf config/in-memory | ${PROCESSOR:-cat} > $TMP_IMC

  local COMMAND_PACKAGE="github.com/mattmoor/mink/pkg/command"
  local KTX_PKG="github.com/mattmoor/mink/pkg/kontext"
  local BP_PKG="github.com/mattmoor/mink/pkg/builds/buildpacks"
  local KO_PKG="github.com/mattmoor/mink/pkg/builds/ko"

  echo -n "-X '${COMMAND_PACKAGE}.BuildDate=${now}' "
  echo -n "-X ${COMMAND_PACKAGE}.Version=${version} "
  echo -n "-X ${COMMAND_PACKAGE}.GitRevision=${rev} "
  echo -n "-X '${COMMAND_PACKAGE}.CoreReleaseURI=${TMP_CORE}' "
  echo -n "-X '${COMMAND_PACKAGE}.InMemoryReleaseURI=${TMP_IMC}' "
  echo -n "-X ${KTX_PKG}.BaseImageString=$(ko publish ${KOFLAGS:-} --tags ${version} -B ./cmd/kontext-expander) "
  echo -n "-X ${KO_PKG}.KoImageString=$(ko publish ${KOFLAGS:-} --tags ${version} -B github.com/google/ko/cmd/ko) "
  echo -n "-X ${BP_PKG}.PlatformSetupImageString=$(ko publish ${KOFLAGS:-} --tags ${version} -B ./cmd/platform-setup) "
  echo -n "-X ${BP_PKG}.ExtractDigestImageString=$(ko publish ${KOFLAGS:-} --tags ${version} -B ./cmd/extract-digest) "
}
