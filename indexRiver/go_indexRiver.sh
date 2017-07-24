#!/bin/sh

server_bin="./indexRiver"
server_ini="mysqlbeat.yml"
server_log="/var/log/indexRiver.log"

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
