var wsUri = "ws://wstail:58888/ws";
var output;
var ws;
var iConnect = 0;
var iConnectId = 0;

function onOpen(evt) {
	status('connected');
	// writeToScreen("CONNECTED");
}

function onClose(evt) {
	status('disconnected, reconnecting');
	reConnect();
}

function onMessage(evt) {

	var data = false;
	try {
		data = JSON.parse(evt.data)
	} catch (e) {
	}
	if (typeof data != 'object' || !data.cmd) {
		console.log('error message', evt.data);
		return;
	}

	if (data.cmd == 'setPos') {
		setPos(data.id, data.x, data.y);
		return;
	}

	if (data.cmd == 'id') {
		console.log('change id');
		iConnectId = data.id
		status('connected, #' + data.id);
		return;
	}

	// status('error, reconnecting');
	writeToScreen('<span style="color: blue;">RESPONSE: ' + evt.data + '</span>');
}

function setPos(id, x, y) {
	var o = $('#box_' + id);
	if (!o.length) {
		var o = $('<div id="box_' + id + '" class="char">' + id + '</div>')
		$('#white_board').append(o);
	}
	o.css('left', x);
	o.css('top',  y);
}

function onError(evt) {
	// writeToScreen('<span style="color: red;">ERROR:</span> ' + evt.data);
	reConnect();
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

function init(i) {
	if (i && i != iConnect) {
		return;
	}

	status('connecting');
	ws = new WebSocket(wsUri);
	ws.onopen = function(evt) { onOpen(evt) };
	ws.onclose = function(evt) { onClose(evt) };
	ws.onmessage = function(evt) { onMessage(evt) };
	ws.onerror = function(evt) { onError(evt) };
}

function reConnect() {
	var i = ++iConnect
	window.setTimeout(function() {
		init(i);
	}, 3000);
}

function status(s) {
	$('#status').text(s);
}

function sendJson(sCmd, aData) {
	aData.cmd = sCmd;
	var s = JSON.stringify(aData);
	ws.send(s);
}

$(document).ready(function () {
	output = $('#output');
	init();

	fn = function(e){
		var parentOffset = $(this).offset();
		var relX = Math.round(e.pageX - parentOffset.left);
		var relY = Math.round(e.pageY - parentOffset.top);
		$('#position_x').text(relX);
		$('#position_y').text(relY);
		sendJson('setPos', {x: relX, y: relY});
		if (iConnectId) {
			setPos(iConnectId, relX, relY);
		}
	}
	$("#white_board").mousemove(fn);
	$("#white_board").mouseover(fn);
});
