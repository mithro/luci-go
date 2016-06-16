#!/bin/bash
# Copyright 2015 The LUCI Authors. All rights reserved.
# Use of this source code is governed under the Apache License, Version 2.0
# that can be found in the LICENSE file.

set -e

mkdir luci-go
cd luci-go

# Download depot_tools
echo "Getting Chromium depot_tools.."
git clone https://chromium.googlesource.com/chromium/tools/depot_tools.git depot_tools
echo

echo "Fetching the infra build..."
"$PWD/depot_tools/fetch" infra

echo "Creating enter script..."
# Create a bashrc include file
ENTER_SCRIPT=$PWD/enter-env.sh
cat > $ENTER_SCRIPT <<EOF
#!/bin/bash
[[ "\${BASH_SOURCE[0]}" != "\${0}" ]] && SOURCED=1 || SOURCED=0
if [ \$SOURCED = 0 ]; then
	exec bash --init-file $ENTER_SCRIPT
fi

if [ -f ~/.bashrc ]; then . ~/.bashrc; fi

export DEPOT_TOOLS="$PWD/depot_tools"
export PATH="\$DEPOT_TOOLS:\$PATH"
export PS1="[luci-go] \$PS1"

cd $PWD/infra/go
eval \$($PWD/infra/go/env.py)

echo "Entered luci-go setup at '$PWD'"
cd "$PWD/infra/go/src/github.com/luci/luci-go"
EOF
chmod a+x $ENTER_SCRIPT

# Running the env setup for the first time
source $ENTER_SCRIPT

# Output usage instructions
if [ -d ~/bin ]; then
	ln -sf $ENTER_SCRIPT ~/bin/luci-go-enter-env
	if which luci-go-enter-env; then
		echo "Enter the environment by running 'luci-go-enter-env'"
		exit 0
	fi
fi
echo "Enter the environment by running '$PWD/enter-env.sh'"
