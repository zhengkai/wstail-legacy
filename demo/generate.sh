#!/bin/bash

OUTFILE='/tmp/fortune.txt'

while :
do
	COUNTER=0
	while [ $COUNTER -lt 100 ]; do

		echo -e "\n"`date -u +'%Y-%m-%d %H:%M:%S'`" UTC\n" >> $OUTFILE
		/usr/games/fortune literature >> $OUTFILE

		let COUNTER=COUNTER+1

		sleep 1
	done
done
