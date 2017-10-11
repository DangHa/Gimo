
function CreateMemberTableFromJSON(data) {

  // Header
  var col = ["#", "Name", "Country", "Scores"];
  var colJSON = ["Position", "Name", "Country", "Scores"]; // de dong bo voi du lieu JSON

  // Goi den bang can tim
  var table = document.getElementById("memberTable");

  // lam dau bang
  var tr = table.insertRow(-1);                   // TABLE ROW.

  for (var i = 0; i < col.length; i++) {
      var th = document.createElement("th");      // TABLE HEADER.
      th.innerHTML = col[i];
      tr.appendChild(th);
  }

  //Vi tri 0 (it's mine)
  trHost = table.insertRow(-1)

  var findBear = ["00", "ah", "Null", "221"]
  for (var j = 0; j < findBear.length; j++) {
      var tabCell = trHost.insertCell(-1);
      tabCell.innerHTML = findBear[j];
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
  var url = 'http://localhost:8080/member'
  $.getJSON(url, function(data){
    console.log("It Worked!");
    CreateMemberTableFromJSON(data)
    console.log(data);
  })

});
