var Papa = require('papaparse');
var fs = require('fs');

var csv = Papa.unparse([
	{
		"Column 1": "foo",
		"Column 2": "bar"
	},
	{
		"Column 1": "abc",
		"Column 2": "def"
	}
], {header: false});


fs.appendFile('file.csv', csv, function(err) {
  if (err) throw err;
  console.log('file saved');
});

fs.appendFile('file.csv', '\n', function(err) {
  if (err) throw err;
  console.log('file saved');
});


fs.appendFile('file.csv', csv, function(err) {
  if (err) throw err;
  console.log('file saved');
});
