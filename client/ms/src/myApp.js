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

SCREEN_WIDTH = 960
SCREEN_HEIGHT = 640

// no use yet
//var Beam = cc.Class.extend({
//    sprite:null,
//    moveForward: function(dt) {
//	angle = this.sprite.getRotation();
//	if (angle < 0)
//	    angle = 360 + angle;
//	r = 160 * dt;
//	x = r * Math.sin(angle / 180 * Math.PI);
//	y = r * Math.cos(angle / 180 * Math.PI);
//
//	this.sprite.setPosition(this.sprite.getPositionX() + x, this.sprite.getPositionY() + y);
//    },
//});

// player ship
var Ship = cc.Class.extend({
    x:0,
    y:0,
    angle:0,
    id:-1,
    sprite: null,
    move: MOVE_NONE,
    rotate: ROTATE_NONE,
    parent:null,
    label:null,
    bloodbar: null,
    deadcd:0,
    waitcd:0,
    dead:false,
    emitter:null,

    ctor:function() {
    },

    restart:function(self) {
	this.x = 480;
	this.y = 320;
	this.angle = 0;
	this.move = MOVE_NONE;
	this.rotate = ROTATE_NONE;
	this.dead = false;

	this.sprite.setPosition(this.x, this.y);
	this.sprite.setRotation(0);

	this.sprite.setVisible(true);
	if (self) {
	    this.bloodbar.setPercentage(100);
	    this.parent.setKeyboardEnabled(true);
	}
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

    sethp: function(hp) {
	this.bloodbar.setPercentage(hp);
    },

    moveForward: function(dt) {
	angle = this.sprite.getRotation() + 90;
	if (angle >= 360)
	    angle = angle - 360;
	r = 80 * dt;
	x = r * Math.sin(angle / 180 * Math.PI);
	y = r * Math.cos(angle / 180 * Math.PI);

	x = this.sprite.getPositionX() + x
	y = this.sprite.getPositionY() + y

	if (x > SCREEN_WIDTH)
	    x = SCREEN_WIDTH
	else if (x < 0)
	    x = 0

	if (y > SCREEN_HEIGHT)
	    y = SCREEN_HEIGHT
	else if (y < 0)
	    y = 0

	this.sprite.setPosition(x, y);
    },

    moveBackward: function(dt) {
	angle = this.sprite.getRotation() + 90;
	if (angle >= 360)
	    angle = angle - 360;
	r = 80 * dt;
	x = r * Math.sin(angle / 180 * Math.PI);
	y = r * Math.cos(angle / 180 * Math.PI);

	x = this.sprite.getPositionX() - x
	y = this.sprite.getPositionY() - y

	if (x > SCREEN_WIDTH)
	    x = SCREEN_WIDTH
	else if (x < 0)
	    x = 0

	if (y > SCREEN_HEIGHT)
	    y = SCREEN_HEIGHT
	else if (y < 0)
	    y = 0

	this.sprite.setPosition(x, y);
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
    },

    // player shoot
    shootBeam: function(update, beamid) {
	idx = null;
	role = null;
	if (update) {
	    role = ME;
	    idx = ME.getBeam();
	    if (idx == null)
		return;
	} else {
	    role = THEM[this.id];
	    idx = beamid
	}

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

	role.shootBeam(idx, _beam);

	// notify server
	if (update)
	    this.sendShootBeam(idx);
    },

    sendmsupdate: function() {
	if (this.dead)
	    return;

	var obj = {
	    cmd: 2,
	    errcode: 0,
	    seq: 0,
	    userid: ME.name,
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

    sendShootBeam: function(idx) {
	var obj = {
	    cmd: 5,
	    errcode: 0,
	    seq: 0,
	    userid: ME.name,
	    body: {
		x: this.sprite.getPositionX(),
		y: this.sprite.getPositionY(),
		angle: this.sprite.getRotation(),
		move: this.move,
		rotate: this.rotate,
		act: 1,
		beamid: idx
	    }
	}
	var str = JSON.stringify(obj, undefined, 2);
	// console.log("send", str);
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
    },

    sendshiprestart: function() {
	var obj = {
	    cmd: 10,
	    errcode: 0,
	    seq: 0,
	    userid: ME.name,
	    body: {
		x: this.sprite.getPositionX(),
		y: this.sprite.getPositionY(),
		angle: this.sprite.getRotation(),
	    }
	}
	var str = JSON.stringify(obj, undefined, 2);
	// console.log("send", str);
	miniConn.send(str);
    },

    die: function(self) {
	// stop notify server ship status
	this.dead = true;

	if (self) {
	    // stop recv control msg from keyboard
	    this.parent.setKeyboardEnabled(false);

	    // set hp to zero
	    this.bloodbar.setPercentage(0);
	}

	// set ship invisible
	this.sprite.setVisible(false);

	// play dead affect
	this.emitter = cc.ParticleFire.create();
        this.parent.addChild(this.emitter, 10);

        this.emitter.setTexture(cc.TextureCache.getInstance().addImage(s_fire));//.pvr"];
        if (this.emitter.setShapeType)
            this.emitter.setShapeType(cc.PARTICLE_BALL_SHAPE);

        var sourcePos = this.emitter.getSourcePosition();
        if (sourcePos.x === 0 && sourcePos.y === 0)
            this.emitter.setPosition(this.sprite.getPositionX(),
				     this.sprite.getPositionY());

	this.deadcd = 1.5;
    },

    procDeadCD: function(dt) {
	// proc deadcd
	if (this.deadcd > 0) {
	    this.deadcd = this.deadcd - dt;
	    if (this.deadcd < 0) {
		this.deadcd = 0;
		this.emitter.stopSystem();
		this.waitcd = 5;
	    }
	}
	if (this.waitcd > 0) {
	    this.waitcd = this.waitcd - dt;
	    if (this.waitcd < 0) {
		this.waitcd = 0;
		this.restart(this.id == myShip.id);
		this.sendshiprestart();
	    }
	}
    }
});

var myShip = new Ship();
var otherShips = new Array(16);
// save data for others
var THEM = new Array(16);

var GameLayer = cc.Layer.extend({
    isMouseDown:false,
    helloImg:null,
    helloLabel:null,
    circle:null,
    emitter:null,

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

	this.scheduleUpdate();
	this.schedule(this.timeCallback, 0.05);
	this.setKeyboardEnabled(true);
    },

    onEnter: function() {
	this._super();

	// add background
	batch = cc.SpriteBatchNode.create(s_bg, 15);
	for (var i = 0; i < 5; i++) {
	    p = cc.Sprite.createWithTexture(batch.getTexture());
	    p.setPosition(i * 256, 0);
	    batch.addChild(p);
	}
	for (var i = 0; i < 5; i++) {
	    p = cc.Sprite.createWithTexture(batch.getTexture());
	    p.setPosition(i * 256, 256);
	    batch.addChild(p);
	}
	for (var i = 0; i < 5; i++) {
	    p = cc.Sprite.createWithTexture(batch.getTexture());
	    p.setPosition(i * 256, 512);
	    batch.addChild(p);
	}
	this.addChild(batch, 0);

	// register cmd
	miniConn.setCmdCallback(3, this.procUserNotify, this);
	miniConn.setCmdCallback(4, this.procKick, this);
	miniConn.setCmdCallback(5, this.procAction, this);
	miniConn.setCmdCallback(7, this.procStopBeam, this);
	miniConn.setCmdCallback(8, this.procShootBeam, this);
	miniConn.setCmdCallback(9, this.procShipDead, this);
	miniConn.setCmdCallback(10, this.procShipRestart, this);

	this.createSelf(myShip.id);

	// add blood bar
        var bloodbar = cc.ProgressTimer.create(cc.Sprite.create(s_hp));
	bloodbar.setType(cc.PROGRESS_TIMER_TYPE_BAR);
        bloodbar.setMidpoint(cc.p(0, 0));
        bloodbar.setBarChangeRate(cc.p(1, 0));
        this.addChild(bloodbar);
	bloodbar.setAnchorPoint(0, 0);
        bloodbar.setPosition(40, 600);
	bloodbar.setPercentage(100);
        //var to1 = cc.ProgressTo.create(2, 100);
        //bloodbar.runAction(cc.RepeatForever.create(to1));
	myShip.bloodbar = bloodbar;
    },

    createSelf: function(id) {
        var size = cc.Director.getInstance().getWinSize();
	myShip.create(id, ME.name, this, size.width / 2, size.height / 2);
    },

    createOtherShip: function(id) {
        var size = cc.Director.getInstance().getWinSize();
	name = ""
	if (THEM[id] != undefined && THEM[id] != null)
	    name = THEM[id].name
	otherShips[id].create(id, name, this, size.width / 2, size.height / 2);
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

	myShip.procDeadCD(dt);
	for (var i = 0; i < 16; i++) {
	    o = otherShips[i];
	    if (o == undefined || o == null)
		continue;

	    o.procDeadCD(dt);
	}
    },

    timeCallback: function(dt) {
	myShip.sendmsupdate();
    },

    procUserNotify: function(target, obj) {
	if (obj.body.users == null) {
	    console.log("usernotify", obj.body);
	}

	for (var i = 0; i < obj.body.users.length; i++) {
	    s = obj.body.users[i];
	    if (s.id == myShip.id) {
		myShip.sethp(s.hp);
		if (s.hp == 0) {
		    console.log("hp is 0");
		}
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

    procShootBeam: function(target, obj) {
	console.log("procShootBeam");
	s = obj.body.data;

	o = otherShips[s.id];
	if (o == undefined || o == null) {
	    otherShips[s.id] = new Ship();
	    otherShips[s.id].setid(s.id);
	    console.log("create other ship", s.id);
	    myShip.parent.createOtherShip(s.id);
	}
	otherShips[s.id].setPos(s.x, s.y, s.angle);
	otherShips[s.id].setMove(s.move, s.rotate);

	o.shootBeam(false, s.beamid);
    },

    procKick: function(target, obj) {
	target.removeOtherShip(obj.body.id);
    },

    procStopBeam: function(target, obj) {
	if (obj.body.data == null)
	    return;

	id = obj.body.data.id;
	beamid = obj.body.data.beamid;

	console.log("stopbeam", obj.body.data)

	if (id == ME.id) {
	    ME.clearBeam(beamid, false);
	    return;
	}

	console.log("stop others beam");
	o = THEM[id];
	if (o == undefined || o == null)
	    return;

	console.log("clear beam");
	o.clearBeam(beamid, false);
    },

    procShipDead: function(target, obj) {
	if (obj.body.data == null)
	    return;

	id = obj.body.data;
	console.log("shipdead", obj);
	console.log("shipdead", obj.body);
	console.log("shipdead", obj.body.data);

	if (id == myShip.id) {
	    myShip.die(true);
	    return;
	}

	o = otherShips[id];
	if (o == undefined || o == null)
	    return;

	o.die(false);
    },

    procShipRestart: function(target, obj) {
	if (obj.body.data == null)
	    return;

	id = obj.body.data;
	console.log("shiprestart", id);
	if (id == myShip.id) {
	    return;
	}

	o = otherShips[id];
	if (o == undefined || o == null)
	    return;

	o.restart(false);
    }
});

var GameScene = cc.Scene.extend({
    onEnter:function () {
        this._super();
	var layer = new GameLayer();
        this.addChild(layer);
        layer.init();
	myShip.setLayer(layer);
    }
});
