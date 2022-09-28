## Install the release package into the system
#!/bin/sh
# darwin-amd64
#   |- cofx
#   |- install.sh
#   |- flowls/

cd $(dirname "$0")

INSTALL_DIR="/usr/local/cofx"
INSTALL_FLOWLS_DIR="/usr/local/cofx/flowls"

if [ ! -d "${INSTALL_DIR}" ]; then 
    mkdir -p ${INSTALL_DIR}
fi
if [ ! -d "${INSTALL_FLOWLS_DIR}" ]; then
    mkdir -p ${INSTALL_FLOWLS_DIR}
fi

cp -f   ./cofx      ${INSTALL_DIR}
cp -rf  ./flowls/*  ${INSTALL_FLOWLS_DIR}

ln -sf ${INSTALL_DIR}/cofx /usr/local/bin/cofx