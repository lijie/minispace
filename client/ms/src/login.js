// Copyright (c) 2014 
// LiJie 2014-05-30 15:44:22
//
// Login Scene
//

var TEXT_INPUT_FONT_NAME = "Arial";
var TEXT_INPUT_FONT_SIZE = 36;

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
    },

    loginCallback: function(target, obj) {
	console.log("login callback", obj, obj.body);
	if (obj.errcode == 0) {
	    myShip.setid(obj.body.id);
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
    }
});

var LoginScene = cc.Scene.extend({
    onEnter:function () {
        this._super();
	var layer = new LoginLayer();
        this.addChild(layer);
        layer.init();
    }
});
