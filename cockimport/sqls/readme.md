

##
## insecure model
# login in insecure

## step 2

aliyun
cockroach sql  --insecure --host=172.31.225.122  < db_product.sql
cockroach sql -u stan --insecure --host=172.31.225.122 < db_product_tables.sql

cockroach sql -u stan --insecure --host=172.31.225.122  -d  db_product
## config hosts
/etc/hosts
172.31.225.122  iZj6cbv0sbzom9p7hjovcuZ

cockroach start --insecure --background --cache=2.5g --port=26257 --http-port=8080 --advertise-host=172.31.225.122 --host=172.31.225.122 --store=path=/root/cockroach-data


## step 3
cockroach sql --insecure --host=10.8.15.167 < root.sql
cockroach sql --insecure --host=10.8.15.167 < db_product.sql
cockroach sql -u stan --insecure --host=10.8.15.167 < db_product_tables.sql


cockroach start --insecure --port=26257 --http-port=8080 --host=10.8.15.167 --store=path=/root/cockroach-data &

cockroach sql -u stan --insecure --host=10.8.15.167 -d db_product

## root
cockroach sql --insecure --host=10.8.15.167 < db_product.sql
### stan
cockroach sql -u stan --insecure --host=10.8.15.167 < db_product_tables.sql

## query count
cockroach sql -u stan --insecure --host=10.8.15.167 -d db_product -e 'select count(*) from t_pro_sell_prices;'


## cockroach quit
cockroach quit --insecure --port=26257 --host=10.8.15.167


## on 96 server

/usr/local/bin/cockroach sql --host=10.8.51.69 --certs-dir=/usr/local/ickey-certs/client-stan --user=stan -d=db_product



cockroach sql -u stan --insecure --host=10.8.15.167 < db_product_tables.sql


cockroach sql --host=10.8.51.69 --certs-dir=/usr/local/ickey-certs/client-stan --user=stan -d=db_product -e 'select count(*) from t_pro_sell_prices;'

cockroach sql --host=10.8.51.69 --certs-dir=/usr/local/ickey-certs/client-stan --user=stan -d=db_product  < db_product.sql
cockroach sql --host=10.8.51.69 --certs-dir=/usr/local/ickey-certs/client-stan --user=stan -d=db_product  < db_product_tables.sql


## Test env nsq startup
nohup nsqlookupd &
nohup nsqd --lookupd-tcp-address=10.8.15.9:4160 -mem-queue-size=100000 &
nohup nsqadmin --lookupd-http-address=10.8.15.9:4161 &
