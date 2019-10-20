

function appendChannel(channel_name){
  var channel_title = document.createElement("div");
  channel_title.innerHTML = channel_name;
  channel_title.className = "channel_title";

  var channel_log = document.createElement("div");
  channel_log.innerHTML = channel_name;
  channel_log.className = "channel_logs";

  channel_title.onclick = function(){selectChannel(channel_title,channel_log);};

  channel_titles.appendChild(channel_title);
  log.appendChild(channel_log);

  channel_logs[channel_name] = channel_log;
};
function selectChannel(c_title,c_log){
  var prev_chl = document.getElementById("selected_channel");
  var prev_log = document.getElementById("channel_log");
  if(prev_chl){
    if(prev_chl.innerHTML!=c_title.innerHTML){
      prev_chl.id = "";
      prev_log.id = "";
      c_title.id = "selected_channel";
      c_log.id = "channel_log";
    }
  } else {
    c_title.id = "selected_channel";
    c_log.id = "channel_log";
  }
};

commands["channel_names"] = function(msg,chl,user){
    var messages = msg.split('::');
    for (var i = 0; i < messages.length; i++) {
      appendChannel(messages[i]);
      appendChannel(messages[i]);
    }
    selectChannel(channel_titles.firstChild,log.firstChild);
};
