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

KEY_UP = 87
KEY_DOWN = 83
KEY_LEFT = 65
KEY_RIGHT = 68

var Conn = cc.Class.extend({
    socket:null,

    ctor:function() {
	socket = new WebSocket("ws://127.0.0.1:12345/minispace")
	socket.onopen = function(e) {}
	socket.onclose = function(e) {}
	socket.onerror = function(e) {}
	socket.onmessage = function(e) {
	    var obj = JSON.parse(e.data)
	    console.log("obj", obj)
	    console.log("body", obj.body.users[0])
	}
	this.socket = socket
    },

    send:function(str) {
	this.socket.send(str);
    }
});

var MyLayer = cc.Layer.extend({
    isMouseDown:false,
    helloImg:null,
    helloLabel:null,
    circle:null,
    sprite:null,
    ship:null,
    _shipro:0,
    ships:[],
    shipcount:0,
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

        // add "Helloworld" splash screen"
//        this.sprite = cc.Sprite.create(s_HelloWorld);
//        this.sprite.setAnchorPoint(0.5, 0.5);
//        this.sprite.setPosition(size.width / 2, size.height / 2);
//        this.sprite.setScale(size.height/this.sprite.getContentSize().height);
//        this.addChild(this.sprite, 0);

	this.ship = cc.Sprite.create(s_ship);
        this.ship.setAnchorPoint(0.5, 0.5);
        this.ship.setPosition(size.width / 2, size.height / 2);
	this.ship.setScale(0.5);
	this.addChild(this.ship, 1);

	this.scheduleUpdate();
	this.schedule(this.timeCallback, 0.05);
	this.setKeyboardEnabled(true);
    },

    moveForward: function() {
	ro = this.ship.getRotation() + 90;
	if (ro < 0)
	    ro = 360 + ro;
	r = 5;
	x = 5 * Math.sin(ro / 180 * Math.PI);
	y = 5 * Math.cos(ro / 180 * Math.PI);

	// console.log("ro", ro, "x", x, "y", y);
	this.ship.setPosition(this.ship.getPositionX() + x, this.ship.getPositionY() + y);
    },

    moveBackward: function() {
	ro = this.ship.getRotation() + 90;
	if (ro < 0)
	    ro = 360 + ro;
	r = 5;
	x = 5 * Math.sin(ro / 180 * Math.PI);
	y = 5 * Math.cos(ro / 180 * Math.PI);

	// console.log("ro", ro, "x", x, "y", y);
	this.ship.setPosition(this.ship.getPositionX() - x, this.ship.getPositionY() - y);
    },

    onKeyUp: function(key) {
	// console.log("key ", key);
	start_move = false;
    },

    onKeyDown: function(key) {
	// console.log("key ", key);
	if (key == KEY_UP) {
	    start_move = true;
	    this.moveForward();
	} else if (key == KEY_DOWN) {
	    start_move = true;
	    this.moveBackward();
	} else if (key == KEY_RIGHT) {
	    ro = this.ship.getRotation(this._shipro) + 4;
	    if (ro >= 360)
		ro = 0;
	    this.ship.setRotation(ro);
	} else if (key == KEY_LEFT) {
	    ro = this.ship.getRotation() - 4;
	    if (ro < 0)
		ro = 360 + ro;
	    this.ship.setRotation(ro);
	}
    },

    update:function(dt) {
//	this.ship.setRotation(this._shipro);
//	this._shipro = this._shipro + 1
//	if (this._shipro > 356)
//	    this._shipro = 0;
//	if (this._shipro % 10 == 0)
//	    this.sendmsupdate()
    },

    timeCallback: function(dt) {
	this.sendmsupdate();
    },

    // add other player in current scene
    addplayer: function(player) {
    },

    shoot: function() {	
    },

    ishit: function() {
    },

    sendmsupdate: function() {
	var obj = {
	    cmd: 2,
	    errcode: 0,
	    seq: 0,
	    usserid: "lijie",
	    body: {
		x: this.ship.getPositionX(),
		y: this.ship.getPositionY(),
		ro: this.ship.getRotation()
	    }
	}
	var str = JSON.stringify(obj, undefined, 2)
	myConn.send(str)
    }
});

var MyScene = cc.Scene.extend({
    onEnter:function () {
        this._super();
        var layer = new MyLayer();
        this.addChild(layer);
        layer.init();
    }
});

var myConn = new Conn();
