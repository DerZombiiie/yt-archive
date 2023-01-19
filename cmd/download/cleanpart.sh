#! /bin/sh

echo "This deletes all partially downloaded files"
echo "download does not use partial downloads and considers them already done"
echo "If you use this software in a cronjob or service, this should be executed prior"
echo "Also you may want to execute it after aborting to clean"
rm outfiles/*.part
