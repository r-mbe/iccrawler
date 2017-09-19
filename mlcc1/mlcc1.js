var Nightmare = require('nightmare');
var Promise = require("bluebird");

const nightmare = Nightmare({ show: true });

nightmare
  .goto('http://www.mlcc1.com')
  .click('#login_c')
  .type('#login_user_name', '13632927231')
  .type('#login_email_password', '@123456')
  .click('#login_button')
  .wait('#content .tree')
  .click("#content > div:nth-child(3) > a")
  .wait('.modal-content')
  // .click('#cate4')
  .goto('http://www.mlcc1.com/search_simplex.html?searchkey=&flag=4')
  .wait('.search_list')
  .evaluate(() => {
    return Array.from(document.querySelectorAll('.search_list  a.parts_n')).map(a => a.href);
  })
  .end()
  .then( links => { console.log(links); return links })
  .then( links => {

    let data = [];

    return Promise.all(links.map(async link => {
       const nightmare2 = Nightmare({ show: true });

       await nightmare2
        .goto(link)
        .wait('.red_bord')
        .evaluate(() => {
          return document.querySelector('.red_bord').innerText;
        })
        .end()
        .then(res => {
          console.log('detail page' + res);
          data.push(res);
        })
        .catch((err => {
          console.error('Detail search failed', err);
        }))
    }))

    return data;
  })
  .catch((error) => {
    console.error('Search failed:', error);
  });
