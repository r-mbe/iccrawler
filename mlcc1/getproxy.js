var request = require('axios');
var Promise = require('bluebird');


module.exports.getProxyIps = async() => {

  var i = 0;
  var proxyIp;
    for(i=0; i< 25; i++) {
      proxyIp = await getOnceProxyIp();
      if (!proxyIp.found) {
          continue;
      } else {
        break;
      }

    }

    //after 25 retrys then final error.

    if (!proxyIp.found) {
      console.log('after 25 retry final can nott get teh best proxy ip!');
    }
    return proxyIp;
}


async function getOnceProxyIp() {

  var result = { proxyip: {}, found: false };

  try {
      // const url = 'http://tvp.daxiangdaili.com/ip/?' + config.daxiang.params;
      //const url = 'http://tvp.daxiangdaili.com' + '/ip/?tid=555040800113736&num=1&foreign=only';

      //IN FIREWALL
      // const url = 'http://tvp.daxiangdaili.com' + '/ip/?tid=555040800113736&num=1';
      const url = 'http://' + '127.0.0.1' + ':9999/api/v1/getip/1';


      //const url = 'http://tvp.daxiangdaili.com' + '/ip/?tid=555040800113736&num=1&foreign=only&category=2&longlife=10';

      //get ips from daxiang api.
      await request.get(url)
          .then(res => {
              console.log("get proxy ip===" + res.data);
              console.log("get proxy ip===" + JSON.stringify(res.data));
              //let sip = JSON.parse(res.data);
              //console.log("get proxy ip===" + sip);
              if(res.data.bestip) {
                if (res.data.bestip.status== 0) {
                    result.proxyip = res.data.bestip.ip;
                    result.found = true;
                    console.log("get xxx proxy ip===" + JSON.stringify(result));
                 }
              }else {
                  console.log("Err get proxy ip doenst exist res ===" + JSON.stringify(result));
              }

          }).catch(e => {
              console.error('axios with catch err' + e);
          });

  } catch (e) {
      console.log('get proxy ip err.' + e);
  }

    return result;
}
