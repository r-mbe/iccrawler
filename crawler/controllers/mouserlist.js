var Nightmare = require('nightmare');
var Promise = require("bluebird");
let cheerio = require("cheerio");
let moment = require("moment");
var config = require('config');
var proxM = require('./getProxyIps')

let baseurl = 'http://www.szlcsc.com/so/global.html&global_search_keyword=';

//http://www.szlcsc.com/search/global.html&global_search_keyword=2222&global_current_catalog=&search_type=



module.exports.crawler = async function(q) {
    ////////////////////////////////////start module

    console.log('ickey =' + q.keyword);
    if (!(typeof q.keyword !== 'undefined' && q.keyword)) {
        console.log('q.keyword undefined...' + q.keyword);
        return undefined;
    }


    var data = await getByKeyword(q.keyword);
    // console.log('Final round ===crawler result =====' + JSON.stringify(data));
    return await data;
    ///////////////////////////////////before module
}

async function getOwnerRow($, html) {
    let sup = await $(html);
    let res = { steps: [], prices: [] };

    // console.log("parse roooooow data ====" + sup);
    //制造商零件编号  manufacturer part
    let manufacturer_part = sup.find('td').eq(3).find('a').text().trim().replace(/\s+/g, "").replace(/\r\n|\n/g, "");
    let manufacturer = sup.find('td').eq(4).find('a').text().trim().replace(/\s+/g, "").replace(/\r\n|\n/g, "");

    res.manufacturer_part = part;
    res.manufacturer = pro_maf;

    return res;
}


//keyword example STPS20M100SG-TR
async function getAllSuppliers(html) {
    let res = { data: []};
    let $ = await cheerio.load(html);
    let owners = await $('table .SearchResultsTable tr');

    if (!(typeof owners !== 'undefined' && owners)) {
        console.log('owners undefined...' + owners);
        return res;
    }

    //  console.log('get szlcsc table html ==' + owners);

    //get owners
    await Promise.all(owners.toArray().slice(2).map(async owner => {
        //get suppliers
        let row = await $(owner).find('tr').eq(1);

        // console.log('get szlcsc table row html ==' + row);

        //console.log('row html =' + row);
        let detail = await getOwnerRow($, row);
        // console.log("AAAffftterr........row =" + JSON.stringify(rdata));
        res.data.push(await detail);
    }));

    return res;
}



async function getByKeyword(keyword) {

    let result = { status: 1, keyword: keyword };
    //try{
    var startT = moment().unix();
    var durT = 0;
    console.log("Now start nightmare time is" + startT);

    var nightmare = {};
    //get proxy ip from daxiang.

    var proxyIp = await proxM.getProxyIps();

    if (!proxyIp.found) {
        console.log('get da xiang proxy ip error\n');
        proxyIp.proxyip = undefined;
    } else {
      console.log('get daxing proxy ip=' + proxyIp.proxyip.toString());
    }


    ///useing da xing proxy ip
    nightmare = Nightmare({
        switches: {
            // 'proxy-server': '10.8.11.240:8100' // set the proxy server here ...
            //'proxy-server': proxyIp // set the proxy server here ...
            'proxy-server': await proxyIp.proxyip // set the proxy server here ...
        },
        webPreferences: {
            images: false
        },
        show: false
    });


    //not using da xiang proxy ip
    // console.log("Debug mode not undefinedneed proxy.")
    // nightmare = Nightmare({ show: false });

    //http://www.szlcsc.com/search/catalog_603_1_0_1-0-0-3-1_0.html&queryBeginPrice=null&queryEndPrice=null
    // var url = baseurl + keyword + "&global_current_catalog=&search_type=";
    var url = {}
    if (Array.isArray(keyword)) {
        url = keyword[0]
    } else {
       url = keyword;
    }

    console.log("nightmare szcsclist XXX will goto url===.", url);

    await nightmare
        .useragent('Mozilla/5.0 (Windows NT 6.1; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/57.0.2987.110 Safari/537.36')
        .goto(url)
        //.wait('#self_product_list_div table#productTab tbody.product_tbody_cls')
        .wait('.Catalog_right_Details tbody.product_tbody_cls ')
        .wait(1000)
        .evaluate(function() {
            //return document.querySelector('#list1545 td.td-part-number a').innerText.replace(/[^\d\.]/g, '');
            //return document.querySelector('#list1545 tbody').innerText;
            return document.querySelector('.Catalog_right_Details').innerHTML;
        })
        .end()
        .then(async(html) => {
            // console.log('get html==== html =' + html);
            durT = moment().unix() - startT;
            console.log("get rows spendxxx" + durT);

            await getAllSuppliers(html)
                .then(data => {
                    durT = moment().unix() - startT;
                    console.log("nightmare request after parset Html Rows. spendxxx" + durT);
                    // console.log("FFFFFFFFFFFFinall after paserHtml content ===" + JSON.stringify(data));

                    //console.log("nightmare request spendxxx")
                    result.status = 0;
                    result.data = data.data;
                    return result;
                });

        }).catch((e) => {
            console.error(e);
        });

    //   }catch(e) {
    //       console.error(e);
    //       return {status: 1, keyword: keyword, data: undefined};
    //   }

    // console.log("====================after await get result=" + JSON.stringify(result));

    durT = moment().unix() - startT;
    console.log("nightmare request spendxxx" + durT);
    return await result;


}


//tool for map to json convert.
function strMapToObj(strMap) {
    let obj = Object.create(null);
    for (let [k, v] of strMap) {
        // We don’t escape the key '__proto__'
        // which can cause problems on older engines
        obj[k] = v;
    }
    return obj;
}

function objToStrMap(obj) {
    let strMap = new Map();
    for (let k of Object.keys(obj)) {
        strMap.set(k, obj[k]);
    }
    return strMap;
}

function strMapToJson(strMap) {
    return JSON.stringify(strMapToObj(strMap));
}

function jsonToStrMap(jsonStr) {
    return objToStrMap(JSON.parse(jsonStr));
}
