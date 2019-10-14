function checkUsername_(){
  var username_val = username.firstChild.value ;
  if(/^[a-z0-9_-]{3,16}$/igm.test(username_val)){
    document.getElementById("username_status").src = "Success.jpg";
    document.getElementById("username_status").title = 'That username looks good!';
    return true;
  } else {
    if(username_val.length<3){
      document.getElementById("username_status").src = "Pending.jpg";
      document.getElementById("username_status").title = "Usernames must be greater than 3 characters long!";
    } else if(username_val.length>=16){
      document.getElementById("username_status").src = "Fail.jpg";
      document.getElementById("username_status").title = "Usernames must be less than 16 characters long!";
    } else if(/^.*\s.*$/igm.test(username_val)){
      document.getElementById("username_status").src = "Fail.jpg";
      document.getElementById("username_status").title = "Usernames must have no spaces in them!";
    } else if(/^.*[!@#$%^&*+=\(\)\[\]\{\}:;,\.\'\`~<>\/\\].*$/igm.test(username_val)){
      document.getElementById("username_status").src = "Fail.jpg";
      document.getElementById("username_status").title = "Usernames must have no special characters in them!";
    } else {
      document.getElementById("username_status").src = "Fail.jpg";
      document.getElementById("username_status").title = "Your username is not valid!";
    }
    return false;
  }
};
function checkPassword_(){
  return true;
}

function login(username_val){
  document.getElementById("popup").style.display = "none";
  document.getElementById("chat_div").style.display = "block";

  conn.send("{collect_channels}");
  conn.send("{collect_friends}");
  conn.send("{collect_resources}");


}
function logout(){
  document.getElementById("popup").style.display = "block";
  document.getElementById("chat_div").style.display = "none";
}
function signin_() {
    if (checkUsername_()&&checkPassword_()){
      var password = document.getElementById("pass").value;
      conn.send("{attempt_login}"+encrypt_(password+username.firstChild.value));
    }
};
function signup_() {
    if (checkUsername_()&&checkPassword_()){
      var password = document.getElementById("pass").value;
      conn.send("{attempt_signup}"+username+","+encrypt_(password+username.firstChild.value));
    }
};
function attempt_logout(){
  conn.send("{attempt_logout}");
};

function encrypt_(upwd){
  return forge_sha256(upwd);
};

commands["login_successful"] = function(result) {
  while (username.firstChild) {
    username.removeChild(username_elem.firstChild);
  }
  var item = document.createElement("div");
  item.innerHTML = createTextLinks_(result);
  username.appendChild(item);

  login(result);
};
commands["login_failed"] = function(result){
  const status = document.getElementById("account_signin_status");
  while (status.firstChild) {
    status.removeChild(status.firstChild);
  }
  var item = document.createElement("div");
  item.innerHTML = createTextLinks_(result);
  status.appendChild(item);
};
commands["signup_successful"] = function(result){
  while (username.firstChild) {
    username.removeChild(username.firstChild);
  }
  var item = document.createElement("div");
  item.innerHTML = createTextLinks_(result);
  username.appendChild(item);

  login(result);
};
commands["signup_failed"] = function(result){
  const status = document.getElementById("account_signin_status");
  while (status.firstChild) {
    status.removeChild(status.firstChild);
  }
  var item = document.createElement("div");
  item.innerHTML = createTextLinks_(result);
  status.appendChild(item);
};
commands["logout_successful"] = function(result){
  logout();
};
