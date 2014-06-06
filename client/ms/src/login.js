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

        this._btngo = this._uiLayer.getWidgetByName("BtnGo");
        this._btngo.addTouchEventListener(this.touchEvent ,this);

	this.scheduleUpdate();
	this.schedule(this.timeCallback, 5);
    },

    loginCallback: function(obj, result) {
	if (result == 0) {
	    console.log("login ok");
	    return;
	}

	console.log("login error", result);
    },

    textFieldEvent: function (sender, type) {
	console.log("here", type);
    },

    touchEvent: function (sender, type) {
        switch (type) {
            case ccs.TouchEventType.ended:
	    console.log("go");
	    myConn.start(this._inputname.getStringValue(),
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
