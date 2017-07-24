#nsqd

## nsq local use too much file message space.


## nsqlookupd
cd /home/xiyx/nsqserver
mkdir -p /home/xiyx/nsqserver/nsq-data
cd /home/xiyx/nsqserver/nsq-data

nohup nsqlookupd &
nohup nsqd --lookupd-tcp-address=10.8.15.9:4160 -mem-queue-size=100000 &
nohup nsqadmin --lookupd-http-address=10.8.15.9:4161 &


-data-path

cockroach start --insecure --port=26257 --http-port=8080 --host=10.8.15.167 --store=path=/root/cockroach-data
cockroach start --insecure --port=26258 --http-port=8081 --host=10.8.15.167 --store=path=/root/cockroach-data2 --join=10.8.15.167:26257
