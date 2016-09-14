var WsTail = {
	ws: false,
	target: null,
	url: null,
	file: null,
	offset: 0,
	ver: 0,
	connectionId: 0,
	wait: 1,
	create: function(option) {
 		var o = $.extend(true, {}, WsTail);
		o.target = option.target || false;
		if (!o.target) {
			this._errorLog('no target', option);
			return;
		}
		o.file = option.file || false;
		if (!o.file) {
			this._errorLog('no file', option);
			return;
		}
		o.url = option.ws || false;
		if (!o.url) {
			this._errorLog('no ws', option);
			return;
		}
		console.log(option)
		o.parse  = option.parse || false;
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
			$this.ws = null;
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
			s = s.substring(1);
			this.offset += this._utf8length(s);
			if (this.parse) {
				s = this.parse(s)
			}
			oText.html(s);
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
	},
	_utf8length: function(str) {
		var m = encodeURIComponent(str).match(/%[89ABab]/g);
		return str.length + (m ? m.length : 0);
	},
	_errorLog: function(s, o) {
		console.log('WsTail Error: ' + s, "\n", o);
	}
}
