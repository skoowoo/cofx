## Install the release package into the system
#!/bin/sh
# darwin-amd64
#   |- cofx
#   |- install.sh
#   |- flowls/

cd $(dirname "$0")
COFX_HOME="${HOME}/.cofx"
FLOWLS_DIR="${COFX_HOME}/flowls/"

if [ ! -d "${COFX_HOME}" ]; then 
    mkdir -p ${COFX_HOME}
fi
if [ ! -d "${FLOWLS_DIR}" ]; then
    mkdir -p ${FLOWLS_DIR}
fi

cp -f   ./cofx      /usr/local/bin/
cp -rf  ./flowls/*  ${FLOWLS_DIR}
