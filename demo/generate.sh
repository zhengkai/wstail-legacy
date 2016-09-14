#!/bin/bash

OUTFILE='/tmp/fortune.txt'

while :
do
	COUNTER=0
	echo > $OUTFILE
	while [ $COUNTER -lt 100 ]; do

		echo -e "\n"`date -u +'%Y-%m-%d %H:%M:%S'`" UTC\n" >> $OUTFILE
		/usr/games/fortune >> $OUTFILE

		let COUNTER=COUNTER+1

		sleep 2.7
	done
done
