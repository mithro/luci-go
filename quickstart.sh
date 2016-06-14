#!/bin/bash
# Copyright 2015 The LUCI Authors. All rights reserved.
# Use of this source code is governed under the Apache License, Version 2.0
# that can be found in the LICENSE file.

set -e

mkdir luci-go
cd luci-go

# Create a bashrc include file
cat > enter-env.sh <<EOF
#!/bin/bash
[[ "\${BASH_SOURCE[0]}" != "\${0}" ]] && SOURCED=1 || SOURCED=0
if [ \$SOURCED = 0 ]; then
	exec bash --init-file $PWD/enter-env.sh
fi

if [ -f ~/.bashrc ]; then . ~/.bashrc; fi

export GOROOT="$PWD/go"
export GOPATH="$PWD/gocode"
export DEPOT_TOOLS="$PWD/depot_tools"
export PATH="\$GOROOT/bin:\$GOPATH/bin:\$DEPOT_TOOLS:\$PATH"
export PS1="[luci-go] \$PS1"
echo "Entered luci-go setup at '$PWD'"
cd "$PWD/gocode/src/github.com/luci/luci-go"
EOF
chmod a+x enter-env.sh

# Get the go runtime
echo "Downloading go runtime.."
GOVER=1.6.2
GORUN=go$GOVER.linux-amd64.tar.gz
wget -c https://storage.googleapis.com/golang/$GORUN -O /tmp/$GORUN
echo
echo "Install go runtime.."
tar -xf /tmp/$GORUN
echo

# Setup go code working directory
mkdir gocode

# Download depot_tools
echo "Getting Chromium depot_tools.."
git clone https://chromium.googlesource.com/chromium/tools/depot_tools.git depot_tools
echo

# The cd will fail because we haven't gotten the code yet
set +e
source enter-env.sh 2> /dev/null
set -e

# Download useful tools
echo "Getting useful tools.."
go get -v -u golang.org/x/tools/cmd/goimports/... # goimports
go get -v -u github.com/maruel/pre-commit-go/cmd/... # pcg
echo

# Download the actual luci-go code
echo "Getting luci-go code.."
go get -v -u github.com/luci/luci-go/...
echo

# Output usage instructions
if [ -d ~/bin ]; then
	ln -sf $PWD/enter-env.sh ~/bin/luci-go-enter-env
	if which luci-go-enter-env; then
		echo "Enter the environment by running 'luci-go-enter-env'"
		exit 0
	fi
fi
echo "Enter the environment by running '$PWD/enter-env.sh'"
