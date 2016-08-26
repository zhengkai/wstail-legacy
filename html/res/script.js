var wsUri = "ws://wstail:58888/ws";
var output;
var ws;

function onOpen(evt) {
	writeToScreen("CONNECTED");
	doSend("WebSocket rocks");
	doSend("yes rpg");
}

function onClose(evt) {
	writeToScreen("DISCONNECTED");
}

function onMessage(evt) {
	writeToScreen('<span style="color: blue;">RESPONSE: ' + evt.data+'</span>');
	// ws.close();
}

function onError(evt) {
	writeToScreen('<span style="color: red;">ERROR:</span> ' + evt.data);
}

function doSend(message) {
	writeToScreen("SENT: " + message);
	ws.send(message);
}

function writeToScreen(message) {
	var pre = document.createElement("p");
	pre.style.wordWrap = "break-word";
	pre.innerHTML = (new Date()).toISOString().slice(11,23) + ' ' + message;
	output.append(pre);
}

function buttonSend() {
	o = $('#send_text');
	s = o.val();
	if (!s.length) {
		return;
	}
	doSend(s);
	o.val('');
}

$(document).ready(function () {
	output = $('#output');
	ws = new WebSocket(wsUri);
	ws.onopen = function(evt) { onOpen(evt) };
	ws.onclose = function(evt) { onClose(evt) };
	ws.onmessage = function(evt) { onMessage(evt) };
	ws.onerror = function(evt) { onError(evt) };
});

// window.addEventListener("load", init, false);
// init();
