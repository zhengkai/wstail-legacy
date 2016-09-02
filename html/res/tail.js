// var wsUri = 'wss://royal-qa.socialgamenet.com:443/ws?file=/tmp/a.txt';
var WsTail = {
	ws: false,
	target: null,
	init: function(wsUri, oTarget) {
		console.log('ws init');
		console.log(wsUri);

		this.target = oTarget;

		var ws = new WebSocket(wsUri);
		this.ws = ws;

		var $this = this;
		ws.onopen    = function(evt) { $this._wsOpen(evt) };
		ws.onclose   = function(evt) { $this._wsClose(evt) };
		ws.onmessage = function(evt) { $this._wsMessage(evt) };
		ws.onerror   = function(evt) { $this._wsError(evt) };

	},
	_wsOpen: function(e) {
		console.log(e)
	},
	_wsClose: function(e) {
		console.log(e)
	},
	_wsMessage: function(e) {
		var s = e.data;

		if (s.length < 2) {
			return
		}

		var cmd = s.substring(0, 1)

		if (cmd == '>') {
			var oText = $('<span></span>');
			oText.text(s.substring(1));
			this.target.append(oText);
			this.target.scrollTop(this.target[0].scrollHeight);
			return;
		}

		if (cmd == '!') {
			cmdMsg = s.substring(1);
			console.log('cmd msg = ' + cmdMsg);
			if (cmdMsg == 'reset') {
				this.target.html('');
				return;
			}
		}
	},
	_wsError: function(e) {
		console.log(e)
	}
}

$(document).ready(function () {
 	var o = $.extend(true, {}, WsTail);
 	o.init('wss://royal-qa.socialgamenet.com:443/ws?file=/tmp/abc.txt', $('#output'));
 	var x = $.extend(true, {}, WsTail);
 	x.init('wss://royal-qa.socialgamenet.com:443/ws?file=/tmp/php-error.txt', $('#xoutput'));
});
