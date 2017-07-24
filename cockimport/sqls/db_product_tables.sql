-- create db_product database
-- DROP DATABASE IF EXISTS db_product;
-- CREATE DATABASE IF NOT EXISTS db_product;
SET DATABASE = db_product;
SHOW DATABASES;

-- create t_pro_sell_stock table
DROP TABLE IF EXISTS t_pro_sell_stocks;
CREATE TABLE IF NOT EXISTS t_pro_sell_stocks (
  id SERIAL  PRIMARY KEY,
  sku   INT NOT NULL ,
  stock_num  INT DEFAULT 0 ,
  frozen_num INT DEFAULT 0 ,
  virtual_num INT DEFAULT 0 ,
  last_update_time INT,
  luptime DATE,
  timeline TIMESTAMPTZ NOT NULL DEFAULT NOW()
);


-- create t_pro_sell_price table
---  currency_id 币种,兼容国内,1为rmb,2为usd,默认usd
---  price_type 价格类型,1-销售价,2-采购价
DROP TABLE IF EXISTS t_pro_sell_prices;
CREATE TABLE IF NOT EXISTS t_pro_sell_prices (
  id SERIAL  PRIMARY KEY,
  mysql_id INT NOT NULL ,
  pro_sell_id INT ,
  sku   INT NOT NULL ,
  price_type INT DEFAULT 1,
  currency_id  INT DEFAULT 2,
  number1 INT DEFAULT 0 ,
  price1  DECIMAL(15,6) DEFAULT 0.0000000 ,
  number2 INT DEFAULT 0 ,
  price2  DECIMAL(15,6) DEFAULT 0.0000000 ,
  number3 INT DEFAULT 0 ,
  price3  DECIMAL(15,6) DEFAULT 0.0000000 ,
  number4 INT DEFAULT 0 ,
  price4  DECIMAL(15,6) DEFAULT 0.0000000 ,
  number5 INT DEFAULT 0 ,
  price5  DECIMAL(15,6) DEFAULT 0.0000000 ,
  number6 INT DEFAULT 0 ,
  price6  DECIMAL(15,6) DEFAULT 0.0000000 ,
  number7 INT DEFAULT 0 ,
  price7  DECIMAL(15,6) DEFAULT 0.0000000 ,
  number8 INT DEFAULT 0 ,
  price8  DECIMAL(15,6) DEFAULT 0.0000000 ,
  number9 INT DEFAULT 0 ,
  price9  DECIMAL(15,6) DEFAULT 0.0000000 ,
  number10 INT DEFAULT 0 ,
  price10  DECIMAL(15,6) DEFAULT 0.0000000 ,
  status   INT DEFAULT 0 ,
  last_update_time INT DEFAULT 0,
  op_admin_id INT DEFAULT 0 ,
  luptime DATE,
  timeline TIMESTAMPTZ NOT NULL DEFAULT NOW()
);