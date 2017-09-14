#!/bin/ksh

unset IFS

while read LINE
do
        LOGFILE=${1:-logfile}.$(date '+%Y%m%d').log
        print "${LINE}" >> ${LOGFILE}
done
