/**
 * gopush-cluster javascript sdk
 */
(function () {	
	//本地存储，用于存储mid
	var storage = window.localStorage;
	var setMid = function(mid){
		if(window.localStorage){
		   storage.setItem("mid",mid)		
		}
	}
	
	var getMid = function() {
		if(window.localStorage && storage.getItem("mid")){
		   return parseInt(storage.getItem("mid"))
		}
		return 0
	}
	
    var getScript = function (options) {
        // JSONP结构
        var callback = 'callback_' + Math.floor(new Date().getTime() * Math.random()).toString(36);
        var head = document.getElementsByTagName("head")[0];
        var script = document.createElement('script');
        options = options || {};
        GoPushCli[callback] = options.success || function () {
        };
        script.type = 'text/javascript';
        script.charset = 'UTF-8';
        script.onload = script.onreadystatechange = function (_, isAbort) {
            if (isAbort || !script.readyState || /loaded|complete/.test(script.readyState)) {
                script.onload = script.onreadystatechange = null;
                head.removeChild(script);
                script = null;
            }
        };
        //script.src = options.url + ((/\?/).test(options.url) ? '&' : '?') + 'callback=GoPushCli.' + callback;
        script.src = options.url + ((/\?/).test(options.url) ? '&' : '?') + 'cb=GoPushCli.' + callback;
        head.insertBefore(script, head.firstChild);
    };

    var parseJSON = function (data) {
        if (window.JSON && window.JSON.parse) {
            return JSON.parse(data);
        }
        return eval('(' + data + ')');
    };

    var GoPushCli = function (options) {
        // Properties
        this.host = options.host;
        this.port = options.port;
        
        // 构建会话 id： {key}_{browser}_{version}_{rn}@{xx}  
        var tmp = options.key;
        var browser = navigator.appCodeName;//navigator.appName;
        var b_version = navigator.appVersion;
        var version = parseFloat(b_version);
        var rn = Math.round(Math.random() * 999); //三位随机数

        if (tmp.indexOf('@') == -1) {
            tmp = tmp + "_" + browser + "-" + version + "-" + rn;
        } else {
            var start = tmp.substring(0, tmp.lastIndexOf('@'));
            var end = tmp.substring(tmp.lastIndexOf('@'))
            tmp = start + "_" + browser + "-" + version + "-" + rn + end;
        }
        this.type = browser;
        this.key = tmp;
		this.originalKey = options.key;
        this.heartbeat = options.heartbeat || 60;
        this.mid = options.mid || getMid();
        this.pmid = options.pmid || 0;
        this.proto = window.WebSocket ? 1 : 2;
        // Timers
        this.heartbeatTimer = null;
        this.timeoutTimer = null;
        // Status
        this.isGetNode = false;
        this.isHandshake = false;
        this.isDesotry = false;
        // Events
        this.onOpen = options.onOpen || function () {
        };
        this.onError = options.onError || function () {
        };
        this.onClose = options.onClose || function () {
        };
        this.onOnlineMessage = options.onOnlineMessage || function () {
        };
        this.onOfflineMessage = options.onOfflineMessage || function () {
        };
    };

    GoPushCli.prototype.start = function () {
        var that = this;
        getScript({
			//获取订阅节点
            url: 'http://' + that.host + ':' + that.port + '/1/server/get?k=' + that.key + '&p=' + that.proto,
            success: function (json) {
                if (json.ret == 0) {
                    that.isGetNode = true;
                    if (that.proto == 1) {
                        that.initWebSocket(json.data.server.split(':'));
                    } else {
                        // TODO Comet
                        that.onError('浏览器不支持WebSocket');
                    }
                } else {
                    that.onError(json.msg);
                }
            }
        });
    };

    GoPushCli.prototype.initWebSocket = function (node) {
        var that = this;
		//订阅
        that.ws = new ReconnectingWebSocket('ws://' + node[0] + ':' + parseInt(node[1]) + '/sub?key=' + that.key + '&heartbeat=' + that.heartbeat);
        //that.ws = new ReconnectingWebSocket('ws://' + node[0] + ':81/sub?key=' + that.key + '&heartbeat=' + that.heartbeat);
        that.ws.onopen = function () {
            var key = that.key;
            var heartbeatStr = that.heartbeat + '';
            that.getOfflineMessage();
            that.runHeartbeatTask();
            that.onOpen();
        };
        that.ws.onmessage = function (e) {
            var data = e.data;
            if (data[0] == '+') {
                clearTimeout(that.timerOutTimer);
                // console.log('Debug: 响应心跳');
            } else if (data[0] == '-') {
                that.onError('握手协议错误' + data);
            } else {
                var message;
                try {
                    message = parseJSON(data);
                } catch (e) {
                    that.onError('解析返回JSON失败');
                    return;
                }
                if (message.gid == 0) {//gid为消息分组ID（0：表示私信，1：表示公共信息）。
                    if (that.mid < message.mid) {
                        that.mid = message.mid;
						//存储mid
						setMid(that.mid)
                    } else {
                        return;
                    }
                } else {
                    if (that.pmid < message.mid) {
                        that.pmid = message.mid;
                    } else {
                        return;
                    }
                }
                that.onOnlineMessage(message);
            }
        };
        that.ws.onclose = function (e) {
            that.onClose();
            that.isDesotry = true;
            clearInterval(that.heartbeatTimer);
            clearTimeout(that.timerOutTimer);
        };
    };

    GoPushCli.prototype.runHeartbeatTask = function () {
        var that = this;
        that.heartbeatTimer = setInterval(function () {
            that.send('h');
            that.timerOutTimer = setTimeout(function () {
                that.destory();
                that.onError('心跳超时');
            }, (that.heartbeat + 15) * 1000);
            // console.log('Debug: 请求心跳');
        }, that.heartbeat * 1000);
    };

    GoPushCli.prototype.send = function (data) {
        if (this.proto == 1) {
            this.ws.send(data);
        } else {
            // Comet TODO
        }
    };

    GoPushCli.prototype.getOfflineMessage = function () {
        var that = this;
        getScript({
			//获取离线消息
            url: 'http://' + that.host + ':' + that.port + '/1/msg/get?k=' + this.originalKey + '&t=' + that.type + '&m=' + that.mid + '&p=' + that.pmid,
            success: function (json) {
                if (json.ret == 0) {
                    var message;
                    var data = json.data;
					
					//pmsgs为老版本gopush返回的消息（1.0 之前的版本）
                    if (data && data.pmsgs) {
                        for (var i = 0, l = data.pmsgs.length; i < l; ++i) {
                            message = parseJSON(data.pmsgs[i]);
                            if (that.pmid < message.mid) {
                                that.onOfflineMessage(message);
                                that.pmid = message.mid;
                            }
                        }
                    }
                    if (data && data.msgs) {
                        for (var i = 0, l = data.msgs.length; i < l; ++i) {
							message =data.msgs[i];
                            if (that.mid < message.mid) {
                                that.onOfflineMessage(message);
                                that.mid = message.mid;
								//存储mid
								setMid(that.mid)
                            }
                        }
                    }
                } else {
                    that.onError(json.msg);
                }
            }
        });
    };

    GoPushCli.prototype.destory = function () {
        this.ws.close();
    };

    window.GoPushCli = GoPushCli;
})();

// MIT License:
//
// Copyright (c) 2010-2012, Joe Walnes
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

/**
 * This behaves like a WebSocket in every way, except if it fails to connect,
 * or it gets disconnected, it will repeatedly poll until it succesfully connects
 * again.
 *
 * It is API compatible, so when you have:
 *   ws = new WebSocket('ws://....');
 * you can replace with:
 *   ws = new ReconnectingWebSocket('ws://....');
 *
 * The event stream will typically look like:
 *  onconnecting
 *  onopen
 *  onmessage
 *  onmessage
 *  onclose // lost connection
 *  onconnecting
 *  onopen  // sometime later...
 *  onmessage
 *  onmessage
 *  etc... 
 *
 * It is API compatible with the standard WebSocket API.
 *
 * Latest version: https://github.com/joewalnes/reconnecting-websocket/
 * - Joe Walnes
 */
(function (global, factory) {
    if (typeof define === 'function' && define.amd) {
        define([], factory);
    } else if (typeof module !== 'undefined' && module.exports) {
        module.exports = factory();
    } else {
        global.ReconnectingWebSocket = factory();
    }
})(this, function () {

    function ReconnectingWebSocket(url, protocols) {
        protocols = protocols || [];

        // These can be altered by calling code.
        this.debug = false;
        this.reconnectInterval = 1000;
        this.reconnectDecay = 1.5;
        this.reconnectAttempts = 0;
        this.timeoutInterval = 2000;

        var self = this;
        var ws;
        var forcedClose = false;
        var timedOut = false;

        this.url = url;
        this.protocols = protocols;
        this.readyState = WebSocket.CONNECTING;
        this.URL = url; // Public API

        this.onopen = function (event) {
        };

        this.onclose = function (event) {
        };

        this.onconnecting = function (event) {
        };

        this.onmessage = function (event) {
        };

        this.onerror = function (event) {
        };

        function connect(reconnectAttempt) {
            ws = new WebSocket(url, protocols);

            if (!reconnectAttempt)
                self.onconnecting();

            if (self.debug || ReconnectingWebSocket.debugAll) {
                console.debug('ReconnectingWebSocket', 'attempt-connect', url);
            }

            var localWs = ws;
            var timeout = setTimeout(function () {
                if (self.debug || ReconnectingWebSocket.debugAll) {
                    console.debug('ReconnectingWebSocket', 'connection-timeout', url);
                }
                timedOut = true;
                localWs.close();
                timedOut = false;
            }, self.timeoutInterval);

            ws.onopen = function (event) {
                clearTimeout(timeout);
                if (self.debug || ReconnectingWebSocket.debugAll) {
                    console.debug('ReconnectingWebSocket', 'onopen', url);
                }
                self.readyState = WebSocket.OPEN;
                reconnectAttempt = false;
                self.reconnectAttempts = 0;
                self.onopen(event);
            };

            ws.onclose = function (event) {
                clearTimeout(timeout);
                ws = null;
                if (forcedClose) {
                    self.readyState = WebSocket.CLOSED;
                    self.onclose(event);
                } else {
                    self.readyState = WebSocket.CONNECTING;
                    self.onconnecting();
                    if (!reconnectAttempt && !timedOut) {
                        if (self.debug || ReconnectingWebSocket.debugAll) {
                            console.debug('ReconnectingWebSocket', 'onclose', url);
                        }
                        self.onclose(event);
                    }
                    setTimeout(function () {
                        self.reconnectAttempts++;
                        connect(true);
                    }, self.reconnectInterval * Math.pow(self.reconnectDecay, self.reconnectAttempts));
                }
            };
            ws.onmessage = function (event) {
                if (self.debug || ReconnectingWebSocket.debugAll) {
                    console.debug('ReconnectingWebSocket', 'onmessage', url, event.data);
                }
                self.onmessage(event);
            };
            ws.onerror = function (event) {
                if (self.debug || ReconnectingWebSocket.debugAll) {
                    console.debug('ReconnectingWebSocket', 'onerror', url, event);
                }
                self.onerror(event);
            };
        }
        connect(false);

        this.send = function (data) {
            if (ws) {
                if (self.debug || ReconnectingWebSocket.debugAll) {
                    console.debug('ReconnectingWebSocket', 'send', url, data);
                }
                return ws.send(data);
            } else {
                throw 'INVALID_STATE_ERR : Pausing to reconnect websocket';
            }
        };

        this.close = function () {
            forcedClose = true;
            if (ws) {
                ws.close();
            }
        };

        /**
         * Additional public API method to refresh the connection if still open (close, re-open).
         * For example, if the app suspects bad data / missed heart beats, it can try to refresh.
         */
        this.refresh = function () {
            if (ws) {
                ws.close();
            }
        };
    }

    /**
     * Setting this to true is the equivalent of setting all instances of ReconnectingWebSocket.debug to true.
     */
    ReconnectingWebSocket.debugAll = false;

    return ReconnectingWebSocket;
});
