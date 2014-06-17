// Copyright (c) 2014 
// LiJie 2014-05-30 15:44:22
//
// Login Scene
//

var TEXT_INPUT_FONT_NAME = "Arial";
var TEXT_INPUT_FONT_SIZE = 36;

var User = cc.Class.extend({
    name: null,
    score: null,
    beams: null,
    id: null,

    ctor: function() {
	this.beams = new Array(5);
    },

    setid: function(id) {
	this.id = id;
    },

    getBeam: function() {
	for (var i = 0; i < 5; i++) {
	    if (this.beams[i] == undefined ||
		this.beams[i] == null) {
		return i;
	    }
	}
	return null;
    },

    clearBeam: function(idx, hit) {
	if (idx < 0 || idx > 5)
	    return;

	console.log("clearBeam");
	if (this.beams[idx] == undefined ||
	    this.beams[idx] == null)
	    return;

	console.log("clearBeam", idx, hit);
	// stop beam action
	this.beams[idx].stopAllActions();
	this.beams[idx].removeFromParent(true);
	this.beams[idx] = undefined;

	// if hit, show affect
    },

    shootBeam: function(idx, beam) {
	console.log("shootbeam", this.id, idx);
	this.beams[idx] = beam;
    }
});

var LoginLayer = cc.Layer.extend({
    _widget: null,
    _uilayer: null,
    _inputname: null,
    _inputpass: null,
    _btngo: null,

    init:function () {
        this._super();
    },

    onEnter:function() {
	this._super();

        this._uiLayer = ccs.UILayer.create();
        this.addChild(this._uiLayer);

        this._widget = ccs.GUIReader.getInstance().widgetFromJsonFile("MiniSpace/MiniSpace_1.json");
        this._uiLayer.addWidget(this._widget);

        this._inputname = this._uiLayer.getWidgetByName("InputName");
        this._inputname.addEventListenerTextField(this.textFieldEvent, this);
        this._inputpass = this._uiLayer.getWidgetByName("InputPass");
        this._inputpass.addEventListenerTextField(this.textFieldEvent, this);

        this._btngo = this._uiLayer.getWidgetByName("BtnGo");
        this._btngo.addTouchEventListener(this.touchEvent ,this);

	this.scheduleUpdate();
	this.schedule(this.timeCallback, 5);

	miniConn.setCmdCallback(6, this.procAddUser, this);
    },

    loginCallback: function(target, obj) {
	console.log("login callback", obj, obj.body);
	if (obj.errcode == 0) {
	    myShip.setid(obj.body.id);
	    ME.setid(obj.body.id);
	    // login ok, go to gamescene
            var director = cc.Director.getInstance();
	    cc.LoaderScene.preload(g_resources, function () {
		director.replaceScene(new GameScene());
            }, this);
	}
    },

    textfieldevent: function (sender, type) {
	console.log("here", type);
    },

    touchEvent: function (sender, type) {
        switch (type) {
            case ccs.TouchEventType.ended:
	    console.log("go");
	    ME.name = this._inputname.getStringValue();
	    miniConn.start(this._inputname.getStringValue(),
			   this._inputpass.getStringValue(),
			   this.loginCallback, this);
            break;

            case ccs.TouchEventType.began:
            case ccs.TouchEventType.moved:
            case ccs.TouchEventType.canceled:
            default:
            break;
        }
    },

    timeCallback: function(dt) {
	console.log("text:", this._inputname.getStringValue());
    },

    procAddUser: function(target, obj) {
	console.log("adduser");
	console.log("adduser", obj);
	console.log("adduser", obj.body);
	console.log("adduser", obj.body.users);

	if (obj.body.users == null ||
	    obj.body.users == undefined)
	    return;

	for (var i = 0; i < obj.body.users.length; i++) {
	    if (obj.body.users[i].id > 16)
		continue;

	    console.log("adduser", obj.body.users[i]);
	    u = obj.body.users[i];
	    o = THEM[u.id];
	    if (o == undefined || o == null) {
		THEM[u.id] = new User();
	    }
	    console.log("new user join", u.name);
	    THEM[u.id].name = u.name;
	    THEM[u.id].setid(u.id);
	}
    }
});

// save data of me
var ME = new User();

var LoginScene = cc.Scene.extend({
    onEnter:function () {
        this._super();
	var layer = new LoginLayer();
        this.addChild(layer);
        layer.init();
    }
});
