// Copyright (c) 2014 Li Jie
// 2014-06-06 17:25:53

// connection manager
var Conn = cc.Class.extend({
    socket:null,
    status:0,

    ctor:function() {
	this.ontest();
    },

    ontest: function() {
	console.log("ontest");
    },

    onNetMessage: function(e) {
	var obj = JSON.parse(e.data)

	if (status == 0) {
	    status = 1;
	    console.log("id", obj.body.id);
	    myShip.setid(obj.body.id);
	    myShip.parent.createSelf(obj.body.id);
	    return;
	}

	if (obj.cmd == 3) {
	    myConn.msupdate(obj);
	    return;
	}

	if (obj.cmd == 4) {
	    myConn.procKick(obj);
	    return;
	}

	if (obj.cmd == 5) {
	    myConn.procAction(obj);
	    return;
	}
    },

    start:function() {
	socket = new WebSocket("ws://10.20.96.187:12345/minispace")
	socket.onopen = function(e) {
	    var obj = {
		cmd: 1,
		errcode: 0,
		seq: 0,
		userid: "lijie",
	    }
	    var str = JSON.stringify(obj, undefined, 2)
	    socket.send(str)
	}

	socket.onclose = function(e) {}
	socket.onerror = function(e) {}
	socket.onmessage = this.onNetMessage;
	this.socket = socket
    },

    send:function(str) {
	this.socket.send(str);
    },

    msupdate:function(obj) {
	for (var i = 0; i < obj.body.users.length; i++) {
	    s = obj.body.users[i];
	    if (s.id == myShip.id) {
		continue;
	    }

	    o = otherShips[s.id];
	    if (o == undefined || o == null) {
		otherShips[s.id] = new Ship();
		otherShips[s.id].setid(s.id);
		console.log("create other ship", s.id);
		myShip.parent.createOtherShip(s.id);
	    }
	    otherShips[s.id].setPos(s.x, s.y, s.angle);
	    otherShips[s.id].setMove(s.move, s.rotate);

	    if (s.act == 1) {
		console.log("recv act", s.act, "id", s.id);
		otherShips[s.id].shootBeam(false);
	    }
	}
    },

    procAction: function(obj) {
	this.msupdate(obj);
	for (var i = 0; i < obj.body.users.length; i++) {
	    s = obj.body.users[i];
	    if (s.id == myShip.id) {
		continue;
	    }

	    o = otherShips[s.id];
	    o.shootBeam(false);
	}
    },

    procKick:function(obj) {
	myShip.parent.removeOtherShip(obj.body.id);
    }
});

var myConn = new Conn();
