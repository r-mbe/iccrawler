#!/bin/sh
indexRiverPath="/root/go/src/techtoolkit.ickey.cn/indexRiver/";

cd ${indexRiverPath};

file="YukiFullRiver";

if [ -f "${file}" ];then
    rm ${file};
fi

#接收输入的供应商
supplier=`echo ${1} | awk '{print $0}'`;
#接收输入的供应商产品id
proSellIds=`echo ${2} | awk '{print $0}'`;

#如果没有输入供应商, 输出提示信息并退出
if [ "${supplier}" = "" ]; then
    echo 'Please input supplier!';
    exit 1;
fi

#如果没有输入供应商产品id, 输出提示信息并退出
if [ "${proSellIds}" = "" ]; then
    echo 'Please input pro_sell_id!';
    exit 2;
fi

go build YukiFullRiver.go

#for test
./YukiFullRiver -url="http://10.8.15.168:9200" -index="${supplier}.es_v2" -type=product -n=100000 -bulk-size=2000 -proSellIds="${proSellIds}"
