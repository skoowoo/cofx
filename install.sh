## Install the release package into the system
#!/bin/sh
# darwin-amd64
#   |- cofx
#   |- install.sh
#   |- flowls/

cd $(dirname "$0")
cofx_home="~/.cofx"
flowls_dir="~/.cofx/flowls/"

if [ ! -d "${cofx_home}" ]; then 
    mkdir -p ${cofx_home}
fi
if [ ! -d "${flowls_dir}" ]; then
    mkdir -p ${flowls_dir}
fi

cp -f   ./cofx      /usr/local/bin/
cp -rf  ./flowls/*  ${flowls_dir}