

function appendChannel(channel_name){
  var item = document.createElement("div");
  item.innerHTML = channel_name;
  item.className = "channel";
  item.onclick = function(){selectChannel(item);};
  channels.appendChild(item);
};
function selectChannel(elem){
  var chl = document.getElementById("selected_channel");
  if(chl){
    chl.id = "";
  }
  elem.id = "selected_channel";
  console.log("select_channel:"+elem.innerHTML);
};

commands["channel_names"] = function(msg,chl,user){
    var messages = msg.split('::');
    for (var i = 0; i < messages.length; i++) {
      appendChannel(messages[i]+i);
      appendChannel(messages[i]+(i+1));
    }
    selectChannel(channels.firstChild);
};
