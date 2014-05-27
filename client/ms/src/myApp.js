/****************************************************************************
 Copyright (c) 2010-2012 cocos2d-x.org
 Copyright (c) 2008-2010 Ricardo Quesada
 Copyright (c) 2011      Zynga Inc.

 http://www.cocos2d-x.org

 Permission is hereby granted, free of charge, to any person obtaining a copy
 of this software and associated documentation files (the "Software"), to deal
 in the Software without restriction, including without limitation the rights
 to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 copies of the Software, and to permit persons to whom the Software is
 furnished to do so, subject to the following conditions:

 The above copyright notice and this permission notice shall be included in
 all copies or substantial portions of the Software.

 THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
 THE SOFTWARE.
 ****************************************************************************/

// const
KEY_UP = 87
KEY_DOWN = 83
KEY_LEFT = 65
KEY_RIGHT = 68

MOVE_NONE = 0
MOVE_FORWARD = 1
MOVE_BACKWARD = 2

ROTATE_NONE = 0
ROTATE_LEFT = 1
ROTATE_RIGHT = 2

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
	    shipLayer.createSelf(obj.body.id);
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
		shipLayer.createOtherShip(s.id);
	    }
	    otherShips[s.id].setPos(s.x, s.y, s.ro);
	    otherShips[s.id].setMove(s.move, s.rotate);
	}
    },

    procKick:function(obj) {
	shipLayer.removeOtherShip(obj.body.id);
    }
});

var Ship = cc.Class.extend({
    x:0,
    y:0,
    ro:0,
    id:-1,
    sprite: null,
    move: MOVE_NONE,
    rotate: ROTATE_NONE,

    ctor:function() {
    },

    setPos:function(x, y, ro) {
	this.x = x;
	this.y = y;
	this.ro = ro;
    },

    synstep: 0,
    updatePos:function() {
	if (this.synstep % 10 == 0) {
	    this.sprite.setPosition(this.x, this.y);
	    this.sprite.setRotation(this.ro);
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

    getRo:function() {
	return this.ro;
    },

    setid:function(id) {
	this.id = id;
    },

    getid:function() {
	return this.id;
    },

    moveForward: function(dt) {
	ro = this.sprite.getRotation() + 90;
	if (ro < 0)
	    ro = 360 + ro;
	r = 80 * dt;
	x = r * Math.sin(ro / 180 * Math.PI);
	y = r * Math.cos(ro / 180 * Math.PI);

	this.sprite.setPosition(this.sprite.getPositionX() + x, this.sprite.getPositionY() + y);
    },

    moveBackward: function(dt) {
	ro = this.sprite.getRotation() + 90;
	if (ro < 0)
	    ro = 360 + ro;
	r = 80 * dt;
	x = r * Math.sin(ro / 180 * Math.PI);
	y = r * Math.cos(ro / 180 * Math.PI);

	this.sprite.setPosition(this.sprite.getPositionX() - x, this.sprite.getPositionY() - y);
    },

    moveLRotate: function(dt) {
	ro = this.sprite.getRotation() - (80 * dt);
	if (ro < 0)
	    ro = 360 + ro;
	this.sprite.setRotation(ro);
    },

    moveRRotate: function(dt) {
	ro = this.sprite.getRotation() + (80 * dt);
	if (ro >= 360)
	    ro = ro - 360;
	this.sprite.setRotation(ro);
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
		ro: this.sprite.getRotation(),
		move: this.move,
		rotate: this.rotate
	    }
	}
	var str = JSON.stringify(obj, undefined, 2);
	myConn.send(str);
    }
});

var myShip = new Ship();
var otherShips = new Array(16);

var MyLayer = cc.Layer.extend({
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

    createSelf: function(id) {
        var size = cc.Director.getInstance().getWinSize();
	myShip.sprite = cc.Sprite.create("ship-" + (id + 1) + ".png");
        myShip.sprite.setAnchorPoint(0.5, 0.5);
        myShip.sprite.setPosition(size.width / 2, size.height / 2);
	myShip.sprite.setScale(0.5);
	this.addChild(myShip.sprite, 1);
    },

    createOtherShip: function(id) {
        var size = cc.Director.getInstance().getWinSize();
	otherShips[id].sprite = cc.Sprite.create("ship-" + (id + 1) + ".png");
        otherShips[id].sprite.setAnchorPoint(0.5, 0.5);
        otherShips[id].sprite.setPosition(size.width / 2, size.height / 2);
	otherShips[id].sprite.setScale(0.5);
	this.addChild(otherShips[id].sprite, 1);
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
    }
});

var shipLayer = null;

var MyScene = cc.Scene.extend({
    onEnter:function () {
        this._super();
	var layer = new MyLayer();
        this.addChild(layer);
        layer.init();
	shipLayer = layer;
	myConn.start();
    }
});

var myConn = new Conn();
