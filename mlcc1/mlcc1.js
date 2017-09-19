var Nightmare = require('nightmare');
var Promise = require("bluebird");
var cheerio = require("cheerio");
var ProxM = require ('./getproxy.js')


async function getByKeyword(url) {


  let nightmare = undefined;
  var proxyIp = await ProxM.getProxyIps();
  if (!proxyIp.found) {
      console.log('get da xiang proxy ip error\n');
      proxyIp.proxyip = undefined;

      nightmare = Nightmare({ show: true })
  } else {
    console.log('get daxing proxy ip=' + proxyIp.proxyip.toString());
    nightmare = Nightmare({
     switches: {
         // 'proxy-server': '10.8.11.240:8100' // set the proxy server here ...
         //'proxy-server': proxyIp // set the proxy server here ...
         'proxy-server': await proxyIp.proxyip.toString() // set the proxy server here ...
     },
     show: true })
  }

await nightmare
  .goto(url)
  .click('#login_c')
  .type('#login_user_name', '13632927231')
  .type('#login_email_password', '@123456')
  .click('#login_button')
  .wait('#content .tree')
  .click("#content > div:nth-child(3) > a")
  .wait('.modal-content')
  .goto('http://www.mlcc1.com/search_simplex.html?searchkey=&flag=4')
  .evaluate(() => {
    return Array.from(document.querySelectorAll('.search_list  a.parts_n')).map(a => a.href);
  })
  .end()
  .then( links => { console.log(links); return links })
  .then( links => {


     let data = [];
    return  Promise.all(links.map(async link => {
      let nightmare2 ;
      var proxyIp = await ProxM.getProxyIps();
      if (!proxyIp.found) {
          console.log('get da xiang proxy ip error\n');
          proxyIp.proxyip = undefined;

          nightmare2 = Nightmare({ show: true });
      } else {
        console.log('get daxing proxy2222 ip=' + proxyIp.proxyip.toString());
        nightmare2 = Nightmare({
         switches: {
             // 'proxy-server': '10.8.11.240:8100' // set the proxy server here ...
             //'proxy-server': proxyIp // set the proxy server here ...
             'proxy-server': await proxyIp.proxyip.toString() // set the proxy server here ...
         },
         show: true })
      }



       return await nightmare2
        .goto(link)
        .wait('.red_bord')
        .evaluate(() => {
          return document.querySelector('.red_bord').innerText;
        })
        .end()
        .then(async res => {
           console.log('XXXrsvj row.' + res);
          return data.push(res);
        })
        .catch((err => {
          console.error('Detail search  detail failed', err);
        }))
    })).then(rows => {
      console.log('XXX YYY::::----->data .' + rows);
      return data;
    })


  })
  .then(rows => {
      console.log('AAAAdd row detail page into array.' + rows);
      return rows;
    // rowd = await getRows(res)
    //
    // if (rowd ){
    //   console.log("one row rowd=" + rowd);
    //   return data;
    // }
  })
  .catch((error) => {
    console.error('Search failed:', error);
  });

}


async function crawlering() {
  let url = 'http://www.mlcc1.com';

  return await getByKeyword(url);
}

 crawlering()
 .then(fres => {
   console.log('final response=' + fres);
 })
 .catch(err => {
   console.error('final cralering error:',err)
 })

//parse one row

  async function getRows(html) {
      let $ = await cheerio.load(html);
      let rows = $('tbody > tr').toArray();
      let res = [];

      await rows.map(async(row) => {
          r = await parseRow($, row)
              //console.log("=========loopppp row=", JSON.stringify(r));
          res.push(r);
      });

      // console.log("now in parseHtml function get resturl rowsres=", res);
      return res;
  }

  async function parseRow($, row) {
      //remove \r\n and space with str replace

      let hrows = $('tbody > tr').toArray();
      let r = {}

      let part = $(row).eq(0).find('td').eq(1).text().trim().replace(/\s+/g, "").replace(/\r\n|\n/g, "");
      let pro_maf = $(row).eq(1).find('td').eq(1).text().trim().replace(/\s+/g, "").replace(/\r\n|\n/g, "");
      let resistance = $(row).eq(2).find('td').eq(1).text().trim().replace(/\s+/g, "").replace(/\r\n|\n/g, "");
      let tolerance = $(row).eq(2).find('td').eq(1).text().trim().replace(/\s+/g, "").replace(/\r\n|\n/g, "");
      let rsize = $(row).eq(2).find('td').eq(1).text().trim().replace(/\s+/g, "").replace(/\r\n|\n/g, "");
      let temp = $(row).eq(2).find('td').eq(1).text().trim().replace(/\s+/g, "").replace(/\r\n|\n/g, "");
      let pd = $(row).eq(2).find('td').eq(1).text().trim().replace(/\s+/g, "").replace(/\r\n|\n/g, "");

      r.part = part;
      r.pro_maf = pro_maf;
      r.resistance = resistance;
      r.tolerance = tolerance;
      r.rsize = rsize;
      r.temp = temp;
      r.pd = pd;

      return await r;
}
