## Used to pack the release files into a tarball
#!/bin/sh

cd $(dirname "$0")

RELEASE_TMP_DIR=".tmp"

if [ ! -d "bin" ]; then 
    echo "bin directory not found"
fi
if [ ! -d "examples" ]; then
    echo "examples directory not found"
fi
if [ ! -d "${RELEASE_TMP_DIR}" ]; then
    mkdir "${RELEASE_TMP_DIR}"
fi

rm -rf "${RELEASE_TMP_DIR}/*"
cp -rf  bin/*  ${RELEASE_TMP_DIR}

for dir in ${RELEASE_TMP_DIR}/*
do
    if [ -d "${dir}" ]; then
        mkdir -p "${dir}/flowls"
        cp -f   install.sh  "${dir}/"
        cp -rf  examples/*  "${dir}/flowls/"

        base=$(basename "${dir}")
        cd .tmp
        tar zcvf "cofx-${base}.tar.gz" "${base}"
        cd -
    fi
done

echo " "
echo "Packed release files into the tarball: "
for p in $(ls ${RELEASE_TMP_DIR}/*.tar.gz)
do
    echo "  $p"
done
