var WsTail = {
	ws: false,
	target: null,
	url: null,
	file: null,
	offset: 0,
	ver: 0,
	connectionId: 0,
	wait: 1,
	create: function(url, file, oTarget) {
 		var o = $.extend(true, {}, WsTail);
		o.target = oTarget;
		o.file = file;
		o.url = url;
 		o.init();
		return o;
	},
	init: function() {
		var url = this.url + '?file=' + encodeURI(this.file) + '&ver=' + this.ver + '&offset=' + this.offset

		var ws = new WebSocket(url);
		this.ws = ws;

		var $this = this;
		ws.onopen    = function (e) { $this._wsOpen(e) }
		ws.onclose   = function (e) { $this._wsClose(e) }
		ws.onmessage = function (e) { $this._wsMessage(e) }
		ws.onerror   = function (e) { $this._wsError(e) }

	},
	_wsOpen: function(e) {
		this.wait = 1;
	},
	_wsClose: function(e) {
		var $this =  this
		setTimeout(function () {
			$this.init();
		}, this.wait * 1000);
		if (this.wait < 4) {
			this.wait++;
		}
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
			this.offset += (s.length - 1);
			this.target.append(oText);
			this.target.scrollTop(this.target[0].scrollHeight);
			return;
		}

		if (cmd == '!') {
			var lCmd = s.substring(1).split(',', 2)
			cmdMsg = lCmd[0];
			cmdContent = lCmd[1];
			if (cmdMsg == 'reset') {
				if (cmdContent.match(/^ver=(\d+)$/)) {
					this.ver = RegExp.$1 - 0;
				}
				this.offset = 0;
				this.target.html('');
				return;
			}
		}
	},
	_wsError: function(e) {
	}
}
