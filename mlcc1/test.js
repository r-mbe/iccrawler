var Papa = require('papaparse');
var fs = require('fs');

var csv = Papa.unparse([{"part":"WSL0805R1000FEA","pro_maf":"威世|Vishay","resistance":"0.1Ω","tolerance":"±1%","rsize":"0805/2.0*1.2mm","temp":"±75ppm/℃","pd":"1/8W"},{"part":"WSL0805R1000FEA","pro_maf":"威世|Vishay","resistance":"0.1Ω","tolerance":"±1%","rsize":"0805/2.0*1.2mm","temp":"±75ppm/℃","pd":"1/8W"}], {header: false});


fs.appendFile('file.csv', csv, function(err) {
  if (err) throw err;
  console.log('file saved');
});

fs.appendFile('file.csv', '\r\n', function(err) {
  if (err) throw err;
  console.log('file saved');
});
