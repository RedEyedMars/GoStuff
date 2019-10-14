function createTextLinks_(text) {
  return (text || "")
    .replace(/\{([^\{\}]+)\}/ig,  function(match, curl){ return "<"+curl+">"})
    .replace(/\{\{([^\}]+)\}\}/ig,function(match, curl){ return "{"+curl+"}"})
    .replace(/([^\S]|^)((([A-Za-z]{3,9}:(?:\/\/)?)(?:[-;:&=\+\$,\w]+@)?[A-Za-z0-9.-]+|(?:www.|[-;:&=\+\$,\w]+@)[A-Za-z0-9.-]+)((?:\/[\+~%\/.\w-_]*)?\??(?:[-\+=&;%@.\w_]*)#?(?:[\w]*))?)/gi,
    function(match, space, url){
      var hyperlink = url;
      if (!hyperlink.match('^https?:\/\/')) {
        hyperlink = 'http://' + hyperlink;
      }
      if(hyperlink.match('https?:\/\/(www.)?youtu\.?be')){
        var match = /https?:\/\/(www.)?youtu(\.be|be\.com)\/(watch\?v=|embed\/)?([^&]+)(&[^&]+)*/g.exec(hyperlink);
        return space + '<iframe width="560" height="315" src="https://www.youtube.com/embed/'+match[4]+'" frameborder="0" allowfullscreen></iframe>';
      }
      else {
        return space + '<a href="' + hyperlink + '" target="_blank">' + url + '</a>';
      }
    });
};

function appendLog(inner) {
  var item = document.createElement("div");
  var indexOfColon = inner.indexOf('::');
  if(indexOfColon>0){
    var chat_user = inner.substring(0,indexOfColon);
    if(chat_user==username){
      item.className = "other_persons_chat";
    }
    item.innerHTML = inner.substring(indexOfColon+2,inner.length());
    item.title = char_user;
  } else {
    item.innerHTML = inner;
  }
  var doScroll = log.scrollTop > log.scrollHeight - log.clientHeight - 1;

  log.appendChild(item);
  if (doScroll) {
      log.scrollTop = log.scrollHeight - log.clientHeight;
  }
};

function submit_chat() {
  if (!conn) {
      return false;
  }
  if (!msg.value) {
      return false;
  }
  if (msg.value.charAt(0) == '/') {
    var index = msg.value.indexOf(' ');
    if (index != -1) {
      conn.send("{"+msg.value.substring(0,index)+"}"+msg.value.substring(index+1));
    }
    else {
      conn.send("{"+msg.value+"}");
    }

  }
  else {
    conn.send("{chat_msg}"+username+"::"+msg.value);
  }
  msg.value = "";
  return false;
};


commands["chat_msg"] = function(result){
    var messages = result.split('\n');
    for (var i = 0; i < messages.length; i++) {
      appendLog(createTextLinks_(messages[i]));
    }
};
commands["admin_msg"] = function(result){
    var messages = result.split('\n');
    for (var i = 0; i < messages.length; i++) {
      appendLog(createTextLinks_(messages[i]));
    }
};
