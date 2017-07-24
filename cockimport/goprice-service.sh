#!/bin/bash
#
# goprice        Startup script for the goprice
#
# chkconfig: - 88 02
# description: The goindex Server is an efficient and extensible  \
#              server implementing the current HTTP standards.
# processname: goprice
# config: /root/go/src/techtoolkit.ickey.cn/indexRiver/mysqlbeat.yml
#
# Usage of bin/go-price-stock:
#  -config string
#       go-price-stock config file (default "./etc/river.toml")
#  -data_dir string
#       path for proRiver to save data
#  -es_addr string
#       Elasticsearch addr
#  -exec string
#       mysqldump execution path
#  -flavor string
#       flavor: mysql or mariadb
#  -my_addr string
#       MySQL addr
#  -my_pass string
#       MySQL password
#  -my_user string
#       MySQL user
#  -server_id int
#       MySQL server id, as a pseudo slave
#
# for product running script

# only need config file name other param can use toml config file support
source /etc/profile
server_bin="/root/go/src/techtoolkit.ickey.cn/price-stock/bin/go-price-stock"
server_ini="/root/go/src/techtoolkit.ickey.cn/price-stock/etc/river.toml"
server_log="/var/log/go-price-stock"
server_folder="/root/go/src/techtoolkit.ickey.cn/price-stock"

start_server()
{
        pid=`ps aux |grep $server_ini |grep -v grep| awk '{print $2}'`
        if [[ $pid -le 0 ]];then
                cd $server_folder
                make clean
                make
                $server_bin  -config=$server_ini  > $server_log &
#                $server_bin  -config=$server_ini  > $server_log 
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
        rm $server_bin
        make clean
        make
}

clean()
{       
        cd $server_folder
        make clean

}

case $1 in
        "start")
                stop_server
                clean
                rebuild
                start_server
                ;;
        "clean")
                clean
                ;;        
        "make")
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
                echo "usage start|stop|restart|make|clean"
                ;;
esac