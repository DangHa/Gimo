
function CreateCountryTableFromJSON(data) {

        //Header
        var col = ["#", "Country", "Scores"];
        var colJSON = ["Position", "Country", "Scores"];

        // Goi den bang can tim
        var table = document.getElementById("countryTable");

        // lam dau bang
        var tr = table.insertRow(-1);                   // TABLE ROW.

        for (var i = 0; i < col.length; i++) {
            var th = document.createElement("th");      // TABLE HEADER.
            th.innerHTML = col[i];
            tr.appendChild(th);
        }

        // add du lieu vao cac dong
        for (var i = 0; i < data.length; i++) {

            tr = table.insertRow(-1);

            for (var j = 0; j < col.length; j++) {
                var tabCell = tr.insertCell(-1);
                tabCell.innerHTML = data[i][colJSON[j]];
            }
        }

      }

$( document ).ready(function() {
  //Get JSON
  var url = 'http://ec2-52-89-58-97.us-west-2.compute.amazonaws.com:8080/country'
  $.getJSON(url, function(data){
    console.log("It Worked!");
    CreateCountryTableFromJSON(data)
    console.log(data);
  })
});
