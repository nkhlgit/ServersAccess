
function goBack(){
    location.href = "/";
    return false;
}

function editSubmit() {
  $.ajax({
    url: "/editSubmit",
    method: "POST",
    data: $("#edit-form").serialize(),
    success: function(rawData) {
          $('#edit-result').html(rawData);
    }
  });
  return false;
}
