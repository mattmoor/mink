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

  mkdir -p generated/dockerfile/$x

  cat > generated/dockerfile/$x/Dockerfile <<EOF
FROM golang:1.16.7 AS build
COPY . /workspace
WORKDIR /workspace
RUN CGO_ENABLED=0 go build -o /workspace/$(basename $x) $x

FROM ${BASE}
COPY --from=build /workspace/$(basename $x) /ko-app/$(basename $x)
COPY ./cmd/webhook/kodata /var/run/ko
ENV KO_DATA_PATH /var/run/ko
ENV PATH ${PATH}:/ko-app
ENTRYPOINT ["/ko-app/$(basename $x)"]

EOF

done

for cfg in $(find config -type f -name '*.yaml'); do

  mkdir -p generated/dockerfile/$(dirname $cfg)
  sed 's@ko://@dockerfile:///generated/dockerfile/@g' $cfg > generated/dockerfile/$cfg

done
