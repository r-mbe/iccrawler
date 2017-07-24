/usr/local/bin/cockroach sql --certs-dir=/tmp/stan --host=10.8.51.69 --port=26257  --user=stan --database=db_product


cockroach sql \
--certs-dir=certs \
--user=maxroach \
--host=12.345.67.89 \
--port=26257 \
--database=critterdb


cockroach start --certs-dir=/root/certs --port=26257 --http-port=8080 --host=10.8.51.69 --store=path=/database/cockroach-data
cockroach start --certs-dir=/root/certs2 --port=26258 --http-port=8081 --host=10.8.51.69 --join=10.8.51.69:26257 --store=path=/database/cockroach-data2
cockroach start --certs-dir=/root/certs3 --port=26259 --http-port=8082 --host=10.8.51.69 --join=10.8.51.69:26257 --store=path=/database/cockroach-data3


mkdir /root/certs
mkdir /root/certs-my
mkdir /root/certs2
mkdir /root/certs2-my
mkdir /root/certs3
mkdir /root/certs3-my

cockroach cert create-ca \
--certs-dir=/root/certs \
--ca-key=/root/certs-my/ca.key

cockroach cert create-client \
root \
--certs-dir=/root/certs \
--ca-key=/root/certs-my/ca.key

cockroach cert create-node \
10.8.51.69 \
10.8.51.69 \
node1.ickey.cn  \
node1.ickey.cn \
localhost \
127.0.0.1 \
10.8.51.115 \
10.8.51.115 \
in-haproxy.ickey.cn  \
in-haproxy.ickey.cn \
--certs-dir=/root/certs \
--ca-key=/root/certs-my/ca.key

cockroach cert create-node --overwrite \
10.8.51.69 \
10.8.51.69 \
node2.ickey.cn  \
node2.ickey.cn \
localhost \
127.0.0.1 \
10.8.51.115 \
10.8.51.115 \
in-haproxy.ickey.cn  \
in-haproxy.ickey.cn \
--certs-dir=/root/certs2 \
--ca-key=/root/certs-my/ca.key

cockroach cert create-node --overwrite \
10.8.51.69 \
10.8.51.69 \
node2.ickey.cn  \
node2.ickey.cn \
localhost \
127.0.0.1 \
10.8.51.115 \
10.8.51.115 \
in-haproxy.ickey.cn  \
in-haproxy.ickey.cn \
--certs-dir=/root/certs3 \
--ca-key=/root/certs-my/ca.key

cockroach start --background  --certs-dir=/root/certs --port=26257 --http-port=8080 --host=10.8.51.69 --store=path=/database/cockroach-data

cockroach start --background  --certs-dir=/root/certs2 --port=26258 --http-port=8081 --host=10.8.51.69 --join=10.8.51.69:26257 --store=path=/database/cockroach-data2
cockroach start --background  --certs-dir=/root/certs3 --port=26259 --http-port=8082 --host=10.8.51.69 --join=10.8.51.69:26257 --store=path=/database/cockroach-data3


/usr/local/bin/cockroach gen haproxy --certs-dir=/root/certs --host=10.8.51.69 --port=26257
--certs-dir=/root/certs \
--host=10.8.51.69 \
--port=26257

创建用户：
##新用户
/usr/local/bin/cockroach user set stan --certs-dir=/root/certs --password --host=10.8.51.69
#ickey@2017

cockroach sql --host=10.8.51.69 --certs-dir=/root/certs --user=stan

#更新密码：
cockroach user set stan --certs-dir=certs --password

mkdir /tmp/stan

/usr/local/bin/cockroach cert create-client \
stan \
--certs-dir=/tmp/stan \
--ca-key=/root/certs/ca.key

/usr/local/bin/cockroach sql \
--certs-dir=/tmp/stan \
--host=10.8.51.69 \
—-user stan \
—database 

https://www.cockroachlabs.com/docs/stable/create-security-certificates.html#create-the-certificate-and-key-pair-for-a-client

##创建client用户

mkdir /root/certs-client
mkdir /root/certs-client-my

cockroach cert create-client \
stan \
--certs-dir=/root/certs-client \
--ca-key=/root/certs-client-my/ca.key


/usr/local/bin/cockroach sql --host=10.8.51.115 --certs-dir=/root/certs-client --user=stan --password

###导入数据
/usr/local/bin/cockroach sql --host=10.8.51.69 --certs-dir=/root/certs --user=root

 /usr/local/bin/cockroach sql --host=10.8.51.69 --certs-dir=/root/certs-client --user=stan < db_product
 /usr/local/bin/cockroach sql --host=10.8.51.69 --certs-dir=/root/certs-client --user=stan < db_product_tables.sql
