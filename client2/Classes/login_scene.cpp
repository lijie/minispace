#include "login_scene.h"
#include "net_node.h"

class MyConn : public NetConn {
 public:
  void onOpen() {
    sc_->startLogin();
  }

  void onClose() {
    printf("conn closed!\n");
  }

  LoginScene *sc_;
};

class LoginCall : public NetCall {
 public:
  void Proc(Json::Value *value);
  LoginScene *sc_;
};

void LoginCall::Proc(Json::Value *value) {
  int errcode = value->get("errcode", -1).asInt();
  if (errcode == 0) {
    sc_->startPlay();
  }
}

#define LOGIN_INIT 0
#define LOGIN_CONNECT 1

bool LoginScene::init() {
  if (!CCScene::init())
    return false;

  state_ = LOGIN_INIT;
  NetNode::Shared()->init();
  return true;
}

void LoginScene::onEnter() {
  CCScene::onEnter();

  ui_layer_ = UILayer::create();
  Layout *layout = static_cast<Layout*>(GUIReader::shareReader()->widgetFromJsonFile("MiniSpace/MiniSpace_1.json"));
  ui_layer_->addWidget(layout);
  addChild(ui_layer_);

  Button *btn = static_cast<Button*>(UIHelper::seekWidgetByName(layout, "BtnGo"));
  btn->addTouchEventListener(this, toucheventselector(LoginScene::onBtnLogin));

  scheduleUpdate();
}

void LoginScene::startPlay() {
}

void LoginScene::startLogin() {
  Json::Value value;
  value["cmd"] = 1;

  NetNode::Shared()->AddCallback(1, new LoginCall);
  NetNode::Shared()->Send(value);
}

void LoginScene::startConnect() {
  if (state_ != LOGIN_INIT)
    return;

  MyConn *conn = new MyConn;
  conn->sc_ = this;

  NetNode::Shared()->Connect("ws://127.0.0.1:12345/echo", conn);
  state_ = LOGIN_CONNECT;
}

void LoginScene::onBtnLogin(CCObject* sender, TouchEventType type) {
  switch (type) {
    case TOUCH_EVENT_ENDED:
      // CCDirector::sharedDirector()->replaceScene(UISceneManager::sharedUISceneManager()->previousUIScene());
      printf("%s\n", __func__);
      startConnect();
      break;

    default:
      break;
  }
}

void LoginScene::update(float dt) {
}

