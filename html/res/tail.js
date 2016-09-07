var WsTail = {
	ws: false,
	target: null,
	uri: null,
	create: function(wsUri, oTarget) {
 		var o = $.extend(true, {}, WsTail);
		console.log('create target', oTarget);
		o.target = oTarget;
		o.uri = wsUri;
 		o.init();
		return o;
	},
	init: function() {
		console.log('init target', this.target);

		var ws = new WebSocket(this.uri);
		this.ws = ws;

		var $this = this;
		ws.onopen    = function (e) { $this._wsOpen(e) }
		ws.onclose   = function (e) { $this._wsClose(e) }
		ws.onmessage = function (e) { $this._wsMessage(e) }
		ws.onerror   = function (e) { $this._wsError(e) }
	},
	_wsOpen: function(e) {
		console.log('ws open', new Date(), e)
	},
	_wsClose: function(e) {
		console.log('ws close', new Date(), e)
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
			if (cmdMsg == 'reset') {
				this.target.html('');
				return;
			}
		}
	},
	_wsError: function(e) {
		console.log('ws error', new Date(), e)
	}
}
