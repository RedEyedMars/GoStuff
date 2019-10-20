

function appendChannel(channel_name){
  var item = document.createElement("div");
  item.innerHTML = channel_name;
  item.className = "channel";
  channels.appendChild(item);
};

commands["channel_names"] = function(msg,chl,user){
    var messages = msg.split('::');
    for (var i = 0; i < messages.length; i++) {
      appendChannel(messages[i]);
      appendChannel(messages[i]);
    }

};
