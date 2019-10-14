

function appendChannel(channel_name){
  var item = document.createElement("div");
  item.innerHTML = inner.substring(indexOfColon+2,inner.length);
  item.title = chat_user;
  item.className = "channel";
  channels.appendChild(item);
};

commands["channel_names"] = function(result){
    var messages = result.split(';;');
    for (var i = 0; i < messages.length; i++) {
      appendChannel(messages[i]);
    }
};
