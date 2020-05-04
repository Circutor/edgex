#!/bin/sh

bin=$(echo $1)
version=$(echo $2 | cut -d "." -f 1)
workspace=$(pwd)
binpath="$workspace/cmd/edgex/$bin"
permissions="700"
name="$bin.ipk"
confs=('core-data' 'core-command' 'core-metadata' 'export-client' 'export-distro' 'support-logging' 'support-scheduler');

if [ ! -f $binpath ]; then
	echo "Binary does not exists"
	exit 0
fi

## Clean IPK
echo "Clean IPK"
rm -rf "$workspace/*.ipk"

## IPK Description
echo "IPK Description"
cd "$workspace/ipk/control/"
sed -i "s/cir.*/cir$version/" control

## IPK Files
echo "IPK Files"
tar -czf ../control.tar.gz ./
cd "$workspace/ipk/data/"
mkdir -p "$workspace/ipk/data/usr/"
mkdir -p "$workspace/ipk/data/usr/bin/"
cp $binpath "$workspace/ipk/data/usr/bin/"
chmod $permissions "$workspace/ipk/data/usr/bin/$bin"
mkdir -p "$workspace/ipk/data/etc/edgex/"

len=${#confs[@]}
i=0
while [ $i -lt $len ]; do
	conf=${confs[$i]}
	mkdir -p "$workspace/ipk/data/etc/edgex/$conf/"
	cp "$workspace/cmd/$conf/res/configuration.toml" "$workspace/ipk/data/etc/edgex/$conf/configuration.toml"
	chmod $permissions "$workspace/ipk/data/etc/edgex/$conf/configuration.toml"
	let i++
done

tar -cJf ../data.tar.xz ./

## IPK Generation
echo "IPK Generation"
cd "$workspace/ipk/"
ar r $name control.tar.gz data.tar.xz debian-binary
mv $name $workspace

## Clean Workspace
echo "Clean Workspace"
rm -rf "$workspace/ipk/control.tar.gz"
rm -rf "$workspace/ipk/data.tar.xz"
rm -rf "$workspace/ipk/data/etc/edgex"
rm -rf "$workspace/ipk/data/usr"
cd "$workspace/ipk/control/"
sed -i "s/cir.*/cir/" control
