// Copyright (c) 2014 
// LiJie 2014-05-28 15:20:33

// const
KEY_UP = 87
KEY_DOWN = 83
KEY_LEFT = 65
KEY_RIGHT = 68
KEY_SHOOT = 85

MOVE_NONE = 0
MOVE_FORWARD = 1
MOVE_BACKWARD = 2

ROTATE_NONE = 0
ROTATE_LEFT = 1
ROTATE_RIGHT = 2

MAX_BEAMCOUNT = 5

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

// no use yet
var Beam = cc.Class.extend({
    sprite:null,
    moveForward: function(dt) {
	angle = this.sprite.getRotation();
	if (angle < 0)
	    angle = 360 + angle;
	r = 160 * dt;
	x = r * Math.sin(angle / 180 * Math.PI);
	y = r * Math.cos(angle / 180 * Math.PI);

	this.sprite.setPosition(this.sprite.getPositionX() + x, this.sprite.getPositionY() + y);
    },
});

// player ship
var Ship = cc.Class.extend({
    x:0,
    y:0,
    angle:0,
    id:-1,
    sprite: null,
    move: MOVE_NONE,
    rotate: ROTATE_NONE,
    beamCount: 0,
    parent:null,
    label:null,

    ctor:function() {
    },

    create: function(id, name, layer, x, y) {
	this.sprite = cc.Sprite.create("ship-" + (id + 1) + ".png");
        this.sprite.setAnchorPoint(0.5, 0.5);
        this.sprite.setPosition(x, y);
	this.sprite.setScale(0.5);
	layer.addChild(this.sprite, 1);
	this.parent = layer;

	// set name
	var label = cc.LabelTTF.create(name, "Arial", 30);
        label.setAnchorPoint(0, 1);
	this.sprite.addChild(label);
	this.label = label;
    },

    setLayer: function(layer) {
	this.parent = layer;
    },

    setPos:function(x, y, angle) {
	this.x = x;
	this.y = y;
	this.angle = angle;
    },

    synstep: 0,
    updatePos:function() {
	if (this.synstep % 10 == 0) {
	    this.sprite.setPosition(this.x, this.y);
	    this.sprite.setRotation(this.angle);
	}
	this.synstep++;
    },

    setMove:function(m, r) {
	this.move = m;
	this.rotate = r;
    },

    getX:function() {
	return this.x;
    },

    getY:function() {
	return this.y;
    },

    getAngle:function() {
	return this.angle;
    },

    setid:function(id) {
	this.id = id;
    },

    getid:function() {
	return this.id;
    },

    moveForward: function(dt) {
	angle = this.sprite.getRotation() + 90;
	if (angle >= 360)
	    angle = angle - 360;
	r = 80 * dt;
	x = r * Math.sin(angle / 180 * Math.PI);
	y = r * Math.cos(angle / 180 * Math.PI);

	this.sprite.setPosition(this.sprite.getPositionX() + x, this.sprite.getPositionY() + y);
    },

    moveBackward: function(dt) {
	angle = this.sprite.getRotation() + 90;
	if (angle >= 360)
	    angle = angle - 360;
	r = 80 * dt;
	x = r * Math.sin(angle / 180 * Math.PI);
	y = r * Math.cos(angle / 180 * Math.PI);

	this.sprite.setPosition(this.sprite.getPositionX() - x, this.sprite.getPositionY() - y);
    },

    moveLRotate: function(dt) {
	angle = this.sprite.getRotation() - (80 * dt);
	if (angle < 0)
	    angle = 360 + angle;
	this.sprite.setRotation(angle);
    },

    moveRRotate: function(dt) {
	angle = this.sprite.getRotation() + (80 * dt);
	if (angle >= 360)
	    angle = angle - 360;
	this.sprite.setRotation(angle);
    },

    // callback
    removeBeam:function (nodeExecutingAction, data) {
        nodeExecutingAction.removeFromParent(data);
	data.beamCount--;
    },

    shootBeam: function(update) {
	// console.log("shootbeam", this.beamCount)
	if (this.beamCount >= MAX_BEAMCOUNT)
	    return;

	_beam = cc.Sprite.create(s_beam1);
	_beam.setPosition(this.sprite.getPositionX(),
			  this.sprite.getPositionY());
        this.parent.addChild(_beam, 1);

	angle = this.sprite.getRotation() + 90;
	_beam.setRotation(this.sprite.getRotation());

	x = 1000 * Math.sin(angle / 180 * Math.PI);
	y = 1000 * Math.cos(angle / 180 * Math.PI);

	var action = cc.Sequence.create(
	    cc.MoveBy.create(3.0, cc.p(x, y)),
	    cc.CallFunc.create(this.removeBeam, _beam, this));
        _beam.runAction(action);
	this.beamCount++;

	// notify server
	if (update)
	    this.sendactupdate(1);
    },

    sendmsupdate: function() {
	var obj = {
	    cmd: 2,
	    errcode: 0,
	    seq: 0,
	    userid: "lijie",
	    body: {
		x: this.sprite.getPositionX(),
		y: this.sprite.getPositionY(),
		angle: this.sprite.getRotation(),
		move: this.move,
		rotate: this.rotate
	    }
	}
	var str = JSON.stringify(obj, undefined, 2);
	miniConn.send(str);
    },

    sendactupdate: function(action) {
	var obj = {
	    cmd: 5,
	    errcode: 0,
	    seq: 0,
	    userid: "lijie",
	    body: {
		x: this.sprite.getPositionX(),
		y: this.sprite.getPositionY(),
		angle: this.sprite.getRotation(),
		move: this.move,
		rotate: this.rotate,
		act: action
	    }
	}
	var str = JSON.stringify(obj, undefined, 2);
	// console.log("send", str);
	miniConn.send(str);
    }
});

var myShip = new Ship();
var otherShips = new Array(16);

var GameLayer = cc.Layer.extend({
    isMouseDown:false,
    helloImg:null,
    helloLabel:null,
    circle:null,
//  sprite:null,
//  ship:null,
    _shipro:0,
    conn:null,

    init:function () {

        //////////////////////////////
        // 1. super init first
        this._super();

        /////////////////////////////
        // 2. add a menu item with "X" image, which is clicked to quit the program
        //    you may modify it.
        // ask director the window size
        var size = cc.Director.getInstance().getWinSize();

        // add a "close" icon to exit the progress. it's an autorelease object
        var closeItem = cc.MenuItemImage.create(
            s_CloseNormal,
            s_CloseSelected,
            function () {
                cc.log("close");
            },this);
        closeItem.setAnchorPoint(0.5, 0.5);

        var menu = cc.Menu.create(closeItem);
        menu.setPosition(0, 0);
        this.addChild(menu, 1);
        closeItem.setPosition(size.width - 20, 20);

        /////////////////////////////
        // 3. add your codes below...
        // add a label shows "Hello World"
        // create and initialize a label
//        this.helloLabel = cc.LabelTTF.create("Hello World", "Impact", 38);
//        // position the label on the center of the screen
//        this.helloLabel.setPosition(size.width / 2, size.height - 40);
//        // add the label as a child to this layer
//        this.addChild(this.helloLabel, 5);

	this.scheduleUpdate();
	this.schedule(this.timeCallback, 0.05);
	this.setKeyboardEnabled(true);
    },

    onEnter: function() {
	this._super();

	// register cmd
	miniConn.setCmdCallback(3, this.procUserNotify, this);
	miniConn.setCmdCallback(5, this.procAction, this);

	this.createSelf(myShip.id);
    },

    createSelf: function(id) {
        var size = cc.Director.getInstance().getWinSize();
	myShip.create(id, "LiJie", this, size.width / 2, size.height / 2);
    },

    createOtherShip: function(id) {
        var size = cc.Director.getInstance().getWinSize();
	otherShips[id].create(id, "test", this, size.width / 2, size.height / 2);
    },

    removeOtherShip: function(id) {
	if (otherShips[id] == undefined || otherShips[id] == null)
	    return

	s = otherShips[id].sprite;
	s.removeFromParent(true);

	otherShips[id] = null;
    },

    onKeyUp: function(key) {
	// console.log("key ", key);
	if (key == KEY_UP) {
	    myShip.move = MOVE_NONE;
	} else if (key == KEY_DOWN) {
	    myShip.move = MOVE_NONE;
	} else if (key == KEY_LEFT) {
	    myShip.rotate = ROTATE_NONE;
	} else if (key == KEY_RIGHT) {
	    myShip.rotate = ROTATE_NONE;
	}
    },

    onKeyDown: function(key) {
	if (key == KEY_UP) {
	    myShip.move = MOVE_FORWARD;
	} else if (key == KEY_DOWN) {
	    myShip.move = MOVE_BACKWARD;
	} else if (key == KEY_RIGHT) {
	    myShip.rotate = ROTATE_RIGHT;
	} else if (key == KEY_LEFT) {
	    myShip.rotate = ROTATE_LEFT;
	} else if (key == KEY_SHOOT) {
	    myShip.shootBeam(true);
	}
    },

    moveSprite: function(sp, dt, setpos) {
	if (setpos) {
	    sp.updatePos();
	}

	if (sp.move == MOVE_FORWARD) {
	    sp.moveForward(dt);
	}
	if (sp.move == MOVE_BACKWARD) {
	    sp.moveBackward(dt);
	}
	if (sp.rotate == ROTATE_LEFT) {
	    sp.moveLRotate(dt);
	}
	if (sp.rotate == ROTATE_RIGHT) {
	    sp.moveRRotate(dt);
	}
    },

    moveShips: function(dt) {
	// move self
	this.moveSprite(myShip, dt, false);

	// move others
	for (var i = 0; i < otherShips.length; i++) {
	    if (i == myShip.getid())
		continue;
	    if (otherShips[i] == undefined || otherShips[i] == null)
		continue;
	    this.moveSprite(otherShips[i], dt, true);
	}
    },

    update:function(dt) {
	this.moveShips(dt);
    },

    timeCallback: function(dt) {
	myShip.sendmsupdate();
    },

    // add other player in current scene
    addplayer: function(player) {
    },

    shoot: function() {	
    },

    ishit: function() {
    },

    procUserNotify: function(target, obj) {
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

    procAction: function(target, obj) {
	this.procUserNotify(target, obj);
	for (var i = 0; i < obj.body.users.length; i++) {
	    s = obj.body.users[i];
	    if (s.id == myShip.id) {
		continue;
	    }

	    o = otherShips[s.id];
	    o.shootBeam(false);
	}
    }
});

var GameScene = cc.Scene.extend({
    onEnter:function () {
        this._super();
	var layer = new GameLayer();
        this.addChild(layer);
        layer.init();
	myShip.setLayer(layer);
//	myConn.start();
    }
});

var myConn = new Conn();
