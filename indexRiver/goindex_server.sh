#!/bin/bash
#
# goindex        Startup script for the goindex
#
# chkconfig: - 88 01
# description: The goindex Server is an efficient and extensible  \
#              server implementing the current HTTP standards.
# processname: goindex
# config: /root/go/src/techtoolkit.ickey.cn/indexRiver/mysqlbeat.yml
#
source /etc/profile
server_bin="/root/go/src/techtoolkit.ickey.cn/indexRiver/indexRiver"
server_ini="/root/go/src/techtoolkit.ickey.cn/indexRiver/mysqlbeat.yml"
server_log="/var/log/indexRiver.log"
server_folder="/root/go/src/techtoolkit.ickey.cn/indexRiver"

start_server()
{
        pid=`ps aux |grep $server_ini |grep -v grep| awk '{print $2}'`
        if [[ $pid -le 0 ]];then
                $server_bin -c $server_ini  > $server_log &
                echo "start ok"
        else
                echo "started"
        fi
}

stop_server()
{
        pid=`ps aux |grep $server_ini |grep -v grep| awk '{print $2}'`
        if [[ $pid -gt 0 ]];then
                kill -9 $pid
                echo "stop ok"
        else
                echo "stopped"
        fi
}

rebuild()
{
        cd $server_folder
        rm indexRiver
        go build main.go 
        mv main indexRiver
}


case $1 in
        "start")
                start_server
                ;;
        "rebuild")
                rebuild
                ;;
        "stop")
                stop_server
                ;;
        "restart")
                stop_server
                start_server
                ;;
        *)
                echo "usage start|stop|restart|rebuild"
                ;;
esac
