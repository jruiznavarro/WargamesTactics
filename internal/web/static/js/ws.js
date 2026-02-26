// WebSocket client for the AoS battle simulator.
var WS = (function() {
    var conn = null;
    var handlers = {};

    function connect(onOpen) {
        var protocol = location.protocol === 'https:' ? 'wss:' : 'ws:';
        conn = new WebSocket(protocol + '//' + location.host + '/ws');

        conn.onopen = function() {
            console.log('WebSocket connected');
            if (onOpen) onOpen();
        };

        conn.onmessage = function(evt) {
            var msg;
            try {
                msg = JSON.parse(evt.data);
            } catch(e) {
                console.error('Invalid WS message:', evt.data);
                return;
            }
            var handler = handlers[msg.type];
            if (handler) {
                handler(msg.data);
            } else {
                console.log('Unhandled WS message:', msg.type, msg.data);
            }
        };

        conn.onerror = function(err) {
            console.error('WebSocket error:', err);
        };

        conn.onclose = function() {
            console.log('WebSocket closed');
        };
    }

    function send(type, data) {
        if (!conn || conn.readyState !== WebSocket.OPEN) {
            console.error('WebSocket not connected');
            return;
        }
        conn.send(JSON.stringify({ type: type, data: data }));
    }

    function on(type, handler) {
        handlers[type] = handler;
    }

    return {
        connect: connect,
        send: send,
        on: on
    };
})();
