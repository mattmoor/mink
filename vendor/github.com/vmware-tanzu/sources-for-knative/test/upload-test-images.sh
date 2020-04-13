#!/usr/bin/env bash

# Copyright 2020 VMware, Inc.
# SPDX-License-Identifier: Apache-2.0

: ${KO_DOCKER_REPO:?"You must set 'KO_DOCKER_REPO', see DEVELOPMENT.md"}

cat | ko resolve --strict -Bf - <<EOF
images:
- ko://github.com/vmware-tanzu/sources-for-knative/vendor/github.com/vmware/govmomi/govc
- ko://github.com/vmware-tanzu/sources-for-knative/test/test_images/listener
EOF
