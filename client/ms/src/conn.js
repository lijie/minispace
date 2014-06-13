// Copyright (c) 2014 Li Jie
// 2014-06-06 17:25:53


// command table
var Command = cc.Class.extend({
    callback: null,
    arg: null,

    ctor: function(cb, arg) {
	this.callback = cb;
	this.arg = arg;
    }
});

var CommandTable = new Array(16);

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

    setCmdCallback: function(cmd, callback, arg) {
	if (cmd > 16)
	    return;

	c = new Command(callback, arg);
	CommandTable[cmd] = c
    },

    onNetMessage: function(e) {
	var obj = JSON.parse(e.data)

	if (obj.cmd > 16)
	    return

	c = CommandTable[obj.cmd];
	if (c == undefined) {
	    return;
	}

	c.callback(c.arg, obj);
	return;

//	if (status == 0) {
//	    status = 1;
//	    console.log("id", obj.body.id);
//	    myShip.setid(obj.body.id);
//	    myShip.parent.createSelf(obj.body.id);
//	    return;
//	}
//
//	if (obj.cmd == 3) {
//	    miniConn.msupdate(obj);
//	    return;
//	}
//
//	if (obj.cmd == 4) {
//	    miniConn.procKick(obj);
//	    return;
//	}
//
//	if (obj.cmd == 5) {
//	    miniConn.procAction(obj);
//	    return;
//	}
    },

    start:function(name, pass, callback, arg) {
	// set login callback
	this.setCmdCallback(1, callback, arg)

	socket = new WebSocket("ws://10.20.96.187:12345/minispace")
	socket.onopen = function(e) {
	    // send login
	    var obj = {
		cmd: 1,
		errcode: 0,
		seq: 0,
		userid: name,
		body: {
		    password: pass
		}
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

var miniConn = new Conn();
