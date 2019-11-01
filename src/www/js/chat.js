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
function appendLog(inner){
  appendChat(inner,null,null);
};
function appendChat(msg,chl,user) {

  var new_chat = document.createElement("div");

  var channel_log;
  if(chl){
    channel_log = channel_logs[chl];
  } else {
    channel_log = document.getElementById("channel_log");
  }

  //var doScroll = channel_log.scrollTop > channel_log.scrollHeight - channel_log.clientHeight - 1;
  if(user){
    if(user==username.innerHTML){
      new_chat.className = "my_persons_chat";
    } else {
      new_chat.className = "other_persons_chat";
    }
    new_chat.innerHTML = msg;
    new_chat.title = user;
    while(channel_log.lastChild!=null&&channel_log.lastChild.title==user){
      new_chat.innerHTML = channel_log.lastChild.innerHTML + "</br>" + new_chat.innerHTML;
      channel_log.removeChild(channel_log.lastChild);
    }
    channel_log.appendChild(new_chat);
  } else {
    new_chat.innerHTML = msg;
    channel_log.appendChild(new_chat);
  }

  //if (doScroll) {
  //    channel_log.scrollTop = channel_log.scrollHeight - channel_log.clientHeight;
  //}
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
    var chl = document.getElementById("selected_channel");
    if(chl){
      conn.send("{chat_msg::"+chl.innerHTML+";;"+username.innerHTML+"}"+msg.value);
    } else {
      conn.send("{chat_msg;;"+username.innerHTML+"}"+msg.value);
    }
  }
  msg.value = "";
  return false;
};


commands["chat_msg"] = function(msg,chl,user){
    var messages = msg.split('\n');
    for (var i = 0; i < messages.length; i++) {
      appendChat(createTextLinks_(messages[i]),chl,user);
    }
};
commands["admin_msg"] = function(msg,chl,user){
    var messages = msg.split('\n');
    for (var i = 0; i < messages.length; i++) {
      appendChat(createTextLinks_(messages[i]),chl,user);
    }
};
