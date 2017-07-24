#!/bin/sh

icksrv_bin="/usr/local/icksrv/icksrv"
icksrv_ini="/usr/local/icksrv/icksrv.ini"
icksrv_log="/var/log/icksrv/icksrv.log"

start_icksrv()
{
        pid=`ps aux |grep $icksrv_bin |grep -v grep| awk '{print $2}'`
        if [[ $pid -le 0 ]];then
                $icksrv_bin $icksrv_ini  > $icksrv_log &
                echo "start ok"
        else
                echo "started"
        fi
}

stop_icksrv()
{
        pid=`ps aux |grep $icksrv_bin |grep -v grep| awk '{print $2}'`
        if [[ $pid -gt 0 ]];then
                kill -9 $pid
                echo "stop ok"
        else
                echo "stopped"
        fi
}

case $1 in
        "start")
                start_icksrv
                ;;
        "stop")
                stop_icksrv
                ;;
        "restart")
                stop_icksrv
                start_icksrv
                ;;
        *)
                echo "usage start|stop|restart"
                ;;
esac
