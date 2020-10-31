#!/usr/bin/env bash

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

for x in $(find config -type f -name '*.yaml' | xargs grep "ko://" | sed 's@.*ko://@@g' | sed 's/",//g' | sort -u); do

  BASE=$(yq r .ko.yaml "baseImageOverrides[$x]")
  if [[ "${BASE:-null}" = "null" ]]; then
    BASE="gcr.io/distroless/static:nonroot-amd64"
  fi

  mkdir -p generated/buildpacks/$x

  if [[ "$x" =~ "github.com/mattmoor/mink" ]]; then
    cat > generated/buildpacks/$x/overrides.toml <<EOF
[[build.env]]
name = "BP_GO_TARGETS"
value = "$(echo $x | sed 's@^github.com/mattmoor/mink/@./@g')"

EOF
  else
    cat > generated/buildpacks/$x/overrides.toml <<EOF
[[build.env]]
name = "BP_GO_TARGETS"
value = "$(echo $x | sed 's@^@./vendor/@g')"

EOF
  fi

done


for cfg in $(find config -type f -name '*.yaml'); do

  mkdir -p generated/buildpacks/$(dirname $cfg)
  cat $cfg | sed 's@ko://@buildpack:///generated/buildpacks/@g' > generated/buildpacks/$cfg

done
