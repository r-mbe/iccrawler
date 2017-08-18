# 获取时间段价格信息

####  1. 获取时间段价格信息

##### 请求方法：POST

接口路径：http://10.8.50.96/go/getprice/v1

描述： 输入参数 获取一段时间的价格变化的走势时间序列数据，调用此接口

|参数名称|是否必须|类型|缺省值|描述|
|---|---|---|---|---|
|sku | yes  | int  |无 | 商品sku |
| from | yes  | string  |无 | 开始时间|
| to | yes  | string  |无 | 结束时间|


注：输入 参数的 json 例子

~~~json

    {
        "sku": 1001180927,
        "from": "2017-10-01 00:00:00",
        "to": "2017-11-01 00:00:00",
    }

~~~

##### 返回结构


描述： 输入参数 获取一段时间的价格变化的走势时间序列数据，调用此接口

|参数名称|是否必须|类型|缺省值|描述|
|---|---|---|---|---|
|sku | yes  | int  |无 | 商品sku |
| mysql_id| yes  | int  |无 | mysql_id|
| pro_sell_id| yes  |int  |无 | pro_sell_id|
| price_type | yes  | int  |无 | 价格类型,1-销售价,2-采购价|
| currency_id| yes  | int  |无 | 币种,兼容国内,1为rmb,2为usd,默认usd|
| number1| yes  | int  |无 | 数量梯度1|
| price1| yes  | decimal  |无 | 价格梯度1|
| number2| yes  | int  |无 | 数量梯度2|
| price2| yes  | decimal  |无 | 价格梯度2|
| number3| yes  | int  |无 | 数量梯度3|
| price3| yes  | decimal  |无 | 价格梯度3|
| number4| yes  | int  |无 | 数量梯度4|
| price4| yes  | decimal  |无 | 价格梯度4|
| number5| yes  | int  |无 | 数量梯度5|
| price5| yes  | decimal  |无 | 价格梯度5|
| number6| yes  | int  |无 | 数量梯度6|
| price6| yes  | decimal  |无 | 价格梯度6|
| number7| yes  | int  |无 | 数量梯度7|
| price7| yes  | decimal  |无 | 价格梯度7|
| number8| yes  | int  |无 | 数量梯度8|
| price8| yes  | decimal  |无 | 价格梯度8|
| number9| yes  | int  |无 | 数量梯度9|
| price9| yes  | decimal  |无 | 价格梯度9|
| number10| yes  | int  |无 | 数量梯度10|
| price10| yes  | decimal  |无 | 价格梯度10|
| status| yes  | int  |无 | 价格状态,默认为0不启用|
| last_update_time| yes  | int  |无 | last_update_time|
| op_admin_id| yes  | int  |无 | last_update_time|
| luptime| yes  | string  |无 | 更新时间|


正确返回结果

~~~json

{
  "success": true,
  "errorCode": 0,
  "message": "",
  "result": [
     {
        "sku": 1001180927,
        "mysql_id": 3456,
        "pro_sell_id": 3333,
        "price_type": 1,
        "currency_id": 1,
        "number1":100,
        "price1": 34.100234,
        "number2":100,
        "price2": 34.100234,
        "number3":100,
        "price3": 34.100234,
        "number4":100,
        "price4": 34.100234,
        "number5":100,
        "price5": 34.100234,
        "number6":100,
        "price6": 34.100234,
        "number7":100,
        "price7": 34.100234,
        "number8":100,
        "price8": 34.100234,
        "number9":100,
        "price9": 34.100234,
        "number10":100,
        "price10": 34.100234,
        "status": 0,
        "last_update_time": 233,
        "op_admin_id": 2344,
        "luptime": "2017-10-01 00:00:00",
  },
  {
    ...
  }
 ]
}
