#!/bin/bash

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

set -o pipefail

# =================================================

# Name of the plugin
BINARY="mink"
PLUGIN="kn-im"

# Directories containing go code which needs to be formatted
SOURCE_DIRS="cmd pkg"

# Directory which should be compiled
MAIN_SOURCE_DIR="cmd/mink"

# =================================================

# Store for later
if [ -z "$1" ]; then
    ARGS=("")
else
    ARGS=("$@")
fi

set -eu

# Run build
run() {
  # Switch on modules unconditionally
  export GO111MODULE=on

  # Jump into project directory
  pushd $(basedir) >/dev/null 2>&1

  # Print help if requested
  if $(has_flag --help -h); then
    display_help
    exit 0
  fi

  # Default flow
  go_build

  echo "────────────────────────────────────────────"
  ./$BINARY version

  # Install if requested
  if $(has_flag --install -i); then
    cp $BINARY ${GOPATH}/bin/
    mkdir -p ~/.config/kn/plugins
    mv $BINARY ~/.config/kn/$PLUGIN
  fi
}


go_build() {
  echo "🚧 Compile"
  go build -mod=vendor -ldflags "$(build_flags $(basedir))" -o $BINARY "./$MAIN_SOURCE_DIR/..."
}

# Dir where this script is located
basedir() {
    # Default is current directory
    local script=${BASH_SOURCE[0]}

    # Resolve symbolic links
    if [ -L $script ]; then
        if readlink -f $script >/dev/null 2>&1; then
            script=$(readlink -f $script)
        elif readlink $script >/dev/null 2>&1; then
            script=$(readlink $script)
        elif realpath $script >/dev/null 2>&1; then
            script=$(realpath $script)
        else
            echo "ERROR: Cannot resolve symbolic link $script"
            exit 1
        fi
    fi

    local dir=$(dirname "$script")
    local full_dir=$(cd "${dir}/.." && pwd)
    echo ${full_dir}
}

# Checks if a flag is present in the arguments.
has_flag() {
    filters="$@"
    for var in "${ARGS[@]}"; do
        for filter in $filters; do
          if [ "$var" = "$filter" ]; then
              echo 'true'
              return
          fi
        done
    done
    echo 'false'
}

# Spaced fillers needed for certain emojis in certain terminals
S=""
X=""

# Calculate space fixing variables S and X
apply_emoji_fixes() {
  # Temporary fix for iTerm issue https://gitlab.com/gnachman/iterm2/issues/7901
  if [ -n "${ITERM_PROFILE:-}" ]; then
    S=" "
    # This issue has been fixed with iTerm2 3.3.7, so let's check for this
    # We can remove this code altogether if iTerm2 3.3.7 is in common usage everywhere
    if [ -n "${TERM_PROGRAM_VERSION}" ]; then
      args=$(echo $TERM_PROGRAM_VERSION | sed -e 's#[^0-9]*\([0-9]*\)[.]\([0-9]*\)[.]\([0-9]*\)\([0-9A-Za-z-]*\)#\1 \2 \3#')
      expanded=$(printf '%03d%03d%03d' $args)
      if [ $expanded -lt "003003007" ]; then
        X=" "
      fi
    fi
  fi
}

# Display a help message.
display_help() {
    cat <<EOT
Build script for Kn plugin $BINARY

Usage: $(basename $BASH_SOURCE) [... options ...]

with the following options:

-i  --install                 Install the resulting plugin into ~/.kn/plugins.
-h  --help                    Display this help message
    --debug                   Debug information for this script (set -x)

You can add a symbolic link to this build script into your PATH so that it can be
called from everywhere. E.g.:

ln -s $(basedir)/hack/build.sh /usr/local/bin/$BINARY-build.sh

EOT
}

if $(has_flag --debug); then
    export PS4='+($(basename ${BASH_SOURCE[0]}):${LINENO}): ${FUNCNAME[0]:+${FUNCNAME[0]}(): }'
    set -x
fi

# Shared funcs with CI
source $(basedir)/hack/build-flags.sh

# Fixe emoji labels for certain terminals
apply_emoji_fixes

run $*
