#include "login_scene.h"
#include "game_scene.h"
#include "role.h"
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
  CCLOG("%s\n", __FUNCTION__);
  int errcode = value->get("errcode", -1).asInt();
  if (errcode == 0) {
    const Json::Value& body = (*value)["body"];
    int id = body.get("id", -1).asInt();
    if (id != -1) {
      Role * r = Role::CreateSelf(id, "test");
      assert(r != NULL);
      sc_->startPlay();
    }
  }
}

class AddUserCall : public NetCall {
 public:
  void Proc(Json::Value *value);
};

void AddUserCall::Proc(Json::Value *value) {
  const Json::Value& v = *value;

  const Json::Value& body = v["body"];
  const Json::Value& users = body["users"];

  if (users.size() == 0)
    return;

  for (unsigned int i = 0; i < users.size(); i++) {
    const Json::Value& u = users[i];
    Role::Create(u["id"].asInt(), u["name"].asString());
    CCLOG("add user %d %s\n", u["id"].asInt(), u["name"].asString().c_str());
  }
}

#define LOGIN_INIT 0
#define LOGIN_CONNECT 1

bool LoginScene::init() {
  if (!CCScene::init())
    return false;

  state_ = LOGIN_INIT;
  // NetNode::Shared()->init();
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

  name_ = static_cast<TextField*>(UIHelper::seekWidgetByName(layout, "InputName"));
  assert(name_ != NULL);
  name_->addEventListenerTextField(this, textfieldeventselector(LoginScene::InputNameEvent));

  pass_ = static_cast<TextField*>(UIHelper::seekWidgetByName(layout, "InputPass"));
  assert(pass_ != NULL);
  pass_->addEventListenerTextField(this, textfieldeventselector(LoginScene::InputNameEvent));

  scheduleUpdate();
  addChild(NetNode::Shared());
}

void LoginScene::InputNameEvent(CCObject *pSender, TextFiledEventType type) {
}


void LoginScene::startPlay() {
  CCLOG("%s\n", __FUNCTION__);

  // keep netnode
  // TODO: do this in NetNode::init() ?
  NetNode::Shared()->retain();

  // remove from login scene
  NetNode::Shared()->removeFromParent();

  GameScene *sc = GameScene::create();
  CCDirector::sharedDirector()->replaceScene(CCTransitionSlideInB::create(0.5, sc));
}

void LoginScene::startLogin() {
  CCLOG("%s %s %s\n", __FUNCTION__, name_->getStringValue(), pass_->getStringValue());
  Json::Value value;
  value["cmd"] = 1;
  value["userid"] = name_->getStringValue();
  
  Json::Value body;
  body["password"] = pass_->getStringValue();

  value["body"] = body;

  NetNode::Shared()->AddCallback(1, new LoginCall);
  NetNode::Shared()->AddCallback(6, new AddUserCall);
  NetNode::Shared()->Send(value);
}

void LoginScene::startConnect() {
  if (state_ != LOGIN_INIT)
    return;

  MyConn *conn = new MyConn;
  conn->sc_ = this;

  NetNode::Shared()->Connect("ws://10.20.96.187:12345/minispace", conn);
  state_ = LOGIN_CONNECT;
}

void LoginScene::onBtnLogin(CCObject* sender, TouchEventType type) {
  switch (type) {
    case TOUCH_EVENT_ENDED:
      // CCDirector::sharedDirector()->replaceScene(UISceneManager::sharedUISceneManager()->previousUIScene());
      startConnect();
      break;

    default:
      break;
  }
}

void LoginScene::update(float dt) {
  // CCLOG("%s\n", __func__);
}

