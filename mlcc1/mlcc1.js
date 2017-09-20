var Nightmare = require('nightmare');
var Promise = require("bluebird");
var cheerio = require("cheerio");
var ProxM = require ('./getproxy.js');
var Papa = require('papaparse');
var fs = require('fs');

var CsvOutFile = "data.csv"

let isShow = false;

async function getByKeyword(url) {

 let fresult = false;

  let nightmare = undefined;
  var proxyIp = await ProxM.getProxyIps();
  if (!proxyIp.found) {
      console.log('get da xiang proxy ip error\n');
      proxyIp.proxyip = undefined;

      nightmare = Nightmare({ show: isShow })
  } else {
    console.log('get daxing proxy ip=' + proxyIp.proxyip.toString());
    nightmare = Nightmare({
     switches: {
         // 'proxy-server': '10.8.11.240:8100' // set the proxy server here ...
         //'proxy-server': proxyIp // set the proxy server here ...
         'proxy-server': await proxyIp.proxyip.toString() // set the proxy server here ...
     },
     show: isShow })
  }


let homeurl = 'http://www.mlcc1.com';

await nightmare
  .useragent('Mozilla/5.0 (Windows NT 6.1; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/57.0.2987.110 Safari/537.36')
  .goto(homeurl)
  .wait(500)
  .click('#login_c')
  .type('#login_user_name', '13632927231')
  .type('#login_email_password', '@123456')
  .click('#login_button')
  .wait(1000)
  .goto(url)
  .wait('.search_list')
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

          nightmare2 = Nightmare({ show: isShow });
      } else {
        console.log('get daxing proxy2222 ip=' + proxyIp.proxyip.toString());
        nightmare2 = Nightmare({
         switches: {
             // 'proxy-server': '10.8.11.240:8100' // set the proxy server here ...
             //'proxy-server': proxyIp // set the proxy server here ...
             'proxy-server': await proxyIp.proxyip.toString() // set the proxy server here ...
         },
         show: isShow })
      }

      setTimeout(function() {
       console.log('Blah blah blah blah extra-blah');
     }, 300);

       return await nightmare2
        .useragent('Mozilla/5.0 (Windows NT 6.1; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/57.0.2987.110 Safari/537.36')
        .goto(link)
        .wait(500)
        .wait('.red_bord')
        .evaluate(() => {
          return document.querySelector('.red_bord').innerHTML;
        })
        .end()
        .then(async res => {
          //  console.log('XXXrsvj row.' + res);
          return data.push(res);
        })
        .catch((err => {
          console.error('Detail search  detail failed', err);
        }))
    })).then(rows => {
      // console.log('XXX YYY::::----->data .' + rows);
      return data;
    })


  })
  .then(htmls => {
      // console.log('AAAAdd row detail page into html.' + htmls);
      // return rows;
      let drows = [];

      return  Promise.all(htmls.map(async html => {
          // console.log('AAAAdd onew.... onew page into html.' + html);
          setTimeout(function() {
           console.log('Blah blah blah blah extra-blah');
         }, 100);

        var r = await getRows(html);

        // console.log('AAAAdd row onew page into html.' + r);
        console.log('AAAAdd row onew page into html.' + JSON.stringify(r));
         return drows.push(r);

      })).then(rows => {
        console.log('.......all rows final parsed data' + JSON.stringify(rows));
        console.log('.......all drows final parsed data' + JSON.stringify(drows));

        var csv = Papa.unparse(drows, {header: false});

        fs.appendFile(CsvOutFile, csv, function(err) {
          if (err) throw err;
          console.log('file append. saved');

        });




        // var csv = Papa.unparse(drows, "data.csv");
        if (drows.length > 0 ) {
          fresult = true;
        }
        return drows;
      })
    //
  })
  .then(data => {
    console.log("Final all rows data=" + JSON.stringify(data));
  })
  .catch((error) => {
    console.error('Search failed:', error);
  });

   return fresult;
}


async function login(url) {
var result = true ;

  try {
        let nightmare = undefined;
        var proxyIp = await ProxM.getProxyIps();
        if (!proxyIp.found) {
            console.log('get da xiang proxy ip error\n');
            proxyIp.proxyip = undefined;

            nightmare = Nightmare({ show: isShow })
        } else {
          console.log('get daxing proxy ip=' + proxyIp.proxyip.toString());
          nightmare = Nightmare({
           switches: {
               // 'proxy-server': '10.8.11.240:8100' // set the proxy server here ...
               //'proxy-server': proxyIp // set the proxy server here ...
               'proxy-server': await proxyIp.proxyip.toString() // set the proxy server here ...
           },
           show: isShow })
        }

        await nightmare
        .useragent('Mozilla/5.0 (Windows NT 6.1; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/57.0.2987.110 Safari/537.36')
        .goto(url)
        .wait(500)
        .click('#login_c')
        .type('#login_user_name', '13632927231')
        .type('#login_email_password', '@123456')
        .click('#login_button')
        // .wait('#content .tree')
        .end()
        .catch((error) => {
          console.error('Search failed:', error);
          result = false;
        });

  } catch (e) {
      console.log('login catch error err.' + e);
      result = false;
  }

  return result;
}

async function crawlering() {
  // let homeurl = 'http://www.mlcc1.com';

  // var islogin = false;

  // islogin = await login(homeurl);


  // if (!islogin) {
  //   console.log('login error.')
  //   return;
  // }

  //crawlering each
  let baseurl = "http://www.mlcc1.com/search_simplex.html?searchkey=&flag=4";
  let url;
  url = baseurl;

  var nums = [];
  // Promise.mapSeries()(var i=0; i<= 112662; i++ ){

  for (var i=0; i<= 112662; i++ ){
    nums.push(i)
  }

      // return  Promise.mapSeries(links.map(async link => {


    for (let i of nums ) {
      if ( i > 0 ){
        url = baseurl + "&p=" + i.toString();
        console.log("now will cralwer page url= " + url);
      }

      try {
        await getByKeyword(url)
              .then(ok => {
                if ( !ok ) {
                  console.log('page crawler error.')
                  fs.appendFile('crawling.log', url + '\n', function(err) {
                    if (err) throw err;
                    console.log('file saved');
                  });
                 }
              });
        console.log("xxxx now will finishedl  = " + url);
        // });
      } catch (e) {
        console.log('page crawler error.')
        fs.appendFile('crawling.log', url + '\n', function(err) {
          if (err) throw err;
          console.log('file saved');
        });
      }

  }

  console.log("final crawlering....")

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
      let row = $('tbody > tr').toArray();

      r = await parseRow($, row)
              //console.log("=========loopppp row=", JSON.stringify(r));

      // console.log("now in parseHtml function get resturl rowsres=", res);
      return r;
  }

  async function parseRow($, row) {
      //remove \r\n and space with str replace

      let hrows = $('tbody > tr').toArray();
      let r = {}

      let part = $(row).eq(0).find('td').eq(1).text().trim().replace(/\s+/g, "").replace(/\r\n|\n/g, "");
      let pro_maf = $(row).eq(1).find('td').eq(1).text().trim().replace(/\s+/g, "").replace(/\r\n|\n/g, "");
      let resistance = $(row).eq(2).find('td').eq(1).text().trim().replace(/\s+/g, "").replace(/\r\n|\n/g, "");
      let tolerance = $(row).eq(3).find('td').eq(1).text().trim().replace(/\s+/g, "").replace(/\r\n|\n/g, "");
      let rsize = $(row).eq(4).find('td').eq(1).text().trim().replace(/\s+/g, "").replace(/\r\n|\n/g, "");
      let temp = $(row).eq(5).find('td').eq(1).text().trim().replace(/\s+/g, "").replace(/\r\n|\n/g, "");
      let pd = $(row).eq(6).find('td').eq(1).text().trim().replace(/\s+/g, "").replace(/\r\n|\n/g, "");

      r.part = part;
      r.pro_maf = pro_maf;
      r.resistance = resistance;
      r.tolerance = tolerance;
      r.rsize = rsize;
      r.temp = temp;
      r.pd = pd;

      return await r;
}
