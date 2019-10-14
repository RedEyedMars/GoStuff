

commands["channel_names"] = function(result){
    var messages = result.split(';;');
    for (var i = 0; i < messages.length; i++) {
      appendLog(createTextLinks_(messages[i]));
    }
};
