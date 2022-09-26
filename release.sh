## Used to pack the release files into a tarball
#!/bin/sh

cd $(dirname "$0")

if [ ! -d "bin" ]; then 
    echo "bin directory not found"
fi
if [ ! -d "examples" ]; then
    echo "examples directory not found"
fi
if [ ! -d ".tmp/" ]; then
    mkdir ".tmp/"
fi

cp -rf  bin/*  .tmp/
for dir in .tmp/*
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
for p in `ls .tmp/*.tar.gz`
do
    echo "  $p"
done