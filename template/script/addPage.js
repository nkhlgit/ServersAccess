function addSubmit() {
  $.ajax({
    url: "/addSubmit",
    method: "POST",
    data: $("#add-form").serialize(),
    success: function(rawData) {
          $('#add-result').html(rawData);
    }
  });
  return false;
}

function goBack(){
    location.href = "/";
    return false;
}