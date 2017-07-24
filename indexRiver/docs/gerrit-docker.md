
# gerrit 基于docker容器的部署：
 
## 命令行启动
#### 1.  启动mysql容器
```
sudo docker run --name mysql -d -v /gerrit_mysql_data:/var/lib/mysql -e MYSQL_ROOT_PASSWORD=123qwezxc -e MYSQL_DATABASE=reviewdb -e MYSQL_USER=gerrit2 -e MYSQL_PASSWORD=gerrit mysql
 ```
 
#### 2. 启动gerrit容器 
    sudo docker run --name gerrit2 --link mysql:db -d -p 8080:8080 -p 29418:29418 -v /gerrit_volume:/var/gerrit/review_site -e DATABASE_TYPE=mysql -e WEBURL=https://9.186.89.219 -e DB_ENV_MYSQL_DB=reviewdb  -e HTTPD_LISTENURL=proxy-https://*:8080/ openfrontier/gerrit
 
#### 3. 启动nginx_ssl_gerrit，此镜像是根据官方nginx镜像生成，对应的

#### Dockerfile配置如下：
```
    a. mkdir nginx_ssl_gerrit && cd nginx_ssl_gerrit
    b. cat > Dockerfile <<EOF
            FROM nginx
            COPY ./nginx/ /etc/nginx/conf.d/
       EOF
    c. nignx的目录结构如下：
        nginx_ssl_gerrit/
        ├── Dockerfile
        └── nginx
            ├── default.conf  #ssl的相关配置
            ├── gerrit.crt    #自签证书公钥
            └── gerrit.key      #自签证书私钥
 
    d. default.conf 配置如下：
        cat > default.conf <<EOF
            server {
                listen 443;
                server_name 127.0.0.1;
 
                ssl  on;
                ssl_certificate      conf.d/gerrit.crt;
                ssl_certificate_key  conf.d/gerrit.key;
 
                location / {
                    proxy_pass              http://gerrit:8080;
                    proxy_set_header        X-Forwarded-For $remote_addr;
                    proxy_set_header        Host $host;
                }
 
                location /login/ {
                    proxy_pass              http://gerrit:8080;
                    proxy_set_header        X-Forwarded-For $remote_addr;
                    proxy_set_header        Host $host;
                }
            }
        EOF
     
    e. 自签公私钥证书的生成
        openssl req -x509 -days 3650 -subj "/CN=9.186.89.219/" -nodes -newkey rsa:4096 -sha256 -keyout gerrit.key -out gerrit.crt
 
 ```

#### 总结：
####     安装过程中可能遇到的问题：
     
        1. Missing project All-Projects
        解决方法：
            进入msyql容器，先删除reviewdb数据库，然后重新创建reviewdb数据库,最后重启启动gerrit容器即可
 
        2. 由于采用的认证方式是OPEID，可能遇到第一个账户登录后，没有管理员权限（gerrit默认第一个登录的用户为管理员账户）
        解决方法：
            a. 删除reviewdb，然后重新创建；
            b. 删除gerrit启动时生成的所有文件，即/gerrit_volume:/var/gerrit/review_site中的所有文件
 
        3. 使用 ssh -p 29418 docker_lab@localhost gerrit gsql, 报错：fatal: docker_lab does not have "Access Database" capability.
        解决方法：
            http://jingyan.baidu.com/article/046a7b3ea8122ef9c27fa919.html
 
为了方便管理，可以使用docker-compose来管理多个容器：
1. docker-compose.yml配置文件如下：
    a. mkdir gerrit-deploy && cd gerrit-deploy
    cat > docker-compose.yml <<EOF
        nginx:
        image: nginx_ssl_gerrit
        restart: always
        links: 
            - gerrit2:gerrit
        ports:
            - 443:443
 
        gerrit2:
            image: openfrontier/gerrit
            restart: always
            links:
                - mysql:db
            ports:
                - 8080:8080
                - 29418:29418
            volumes:
                - /gerrit_volume:/var/gerrit/review_site
            environment:
                - DATABASE_TYPE=mysql
                - DB_ENV_MYSQL_DB=reviewdb
                - WEBURL=https://9.186.89.219
                - HTTPD_LISTENURL=proxy-https://*:8080/
 
        mysql:
            image: mysql
            restart: always
            volumes:
                - /gerrit_mysql_data:/var/lib/mysql
            environment:
                - MYSQL_ROOT_PASSWORD=123qwezxc
                - MYSQL_DATABASE=reviewdb
                - MYSQL_USER=gerrit2
                - MYSQL_PASSWORD=gerrit
    EOF
 
2. 启动容器:
    sudo docker-compose up -d 
 
3. 查看是否启动：
    sudo docker-compose ps