#include <cocos2d.h>
#include "game_scene.h"
#include "net_node.h"
#include "role.h"
// #include <cocos-ext.h>

USING_NS_CC;
// USING_NS_CC_EXT;

class NPCLayer : public CCLayer {
 public:
  CREATE_FUNC(NPCLayer);
};

class BgLayer : public CCLayer {
 public:
  CREATE_FUNC(BgLayer);

  bool init() {
    CCLayer::init();
    return true;
  }

  void onEnter() {
    CCLayer::onEnter();

    CCSprite *p;
    // add background
    CCSpriteBatchNode *batch = CCSpriteBatchNode::create("background.png", 90);
    for (int i = 0; i < 10; i++) {
      p = CCSprite::createWithTexture(batch->getTexture());
      p->setPosition(ccp(i * 256, 0));
      addChild(p);
    }
    for (int i = 0; i < 10; i++) {
      p = CCSprite::createWithTexture(batch->getTexture());
      p->setPosition(ccp(i * 256, 256));
      addChild(p);
    }
    for (int i = 0; i < 10; i++) {
      p = CCSprite::createWithTexture(batch->getTexture());
      p->setPosition(ccp(i * 256, 512));
      addChild(p);
    }
    for (int i = 0; i < 10; i++) {
      p = CCSprite::createWithTexture(batch->getTexture());
      p->setPosition(ccp(i * 256, 768));
      addChild(p);
    }
    for (int i = 0; i < 10; i++) {
      p = CCSprite::createWithTexture(batch->getTexture());
      p->setPosition(ccp(i * 256, 1024));
      addChild(p);
    }
    for (int i = 0; i < 10; i++) {
      p = CCSprite::createWithTexture(batch->getTexture());
      p->setPosition(ccp(i * 256, 1024 + 256));
      addChild(p);
    }
    for (int i = 0; i < 10; i++) {
      p = CCSprite::createWithTexture(batch->getTexture());
      p->setPosition(ccp(i * 256, 1024 + 256 * 2));
      addChild(p);
    }
    // this.addChild(batch, 0);
  }

 private:
};

class UserNotifyCall : public NetCall {
 public:
  void Proc(Json::Value *value);
  CCNode *npc_;
};

void UserNotifyCall::Proc(Json::Value *value) {
  const Json::Value& v = *value;

  const Json::Value& body = v["body"];
  const Json::Value& users = body["users"];

  if (users.size() == 0)
    return;

  for (unsigned int i = 0; i < users.size(); i++) {
    const Json::Value& u = users[i];
    int id = u["id"].asInt();

    if (id == Role::Self()->id()) {
      // set hp
      continue;
    }

    Role *o = Role::FindByID(id);
    if (o == NULL) {
      CCLOG("user id %d doesnot exist!\n", id);
      continue;
    }

    float x = u["x"].asFloat();
    float y = u["y"].asFloat();
    float destx = u["destx"].asFloat();
    float desty = u["desty"].asFloat();
    float angle = u["angle"].asFloat();
    int move = u["move"].asInt();
    int rotate = u["rotate"].asInt();

    if (o->sprite() == NULL)
      o->Init(npc_, ccp(x, y));
    o->set_loc(ccp(x, y));
    // o->sprite()->setPosition(ccp(x, y));
    o->set_angle(angle);
    o->set_move(move, rotate);
    o->FlushLoc();
    o->set_dest(ccp(destx, desty));
  }

}

class ShootBeamCall : public NetCall {
 public:
  void Proc(Json::Value *value);
};

void ShootBeamCall::Proc(Json::Value *value) {
  const Json::Value& v = *value;

  const Json::Value& body = v["body"];
  const Json::Value& data = body["data"];

  int id = data["id"].asInt();
  Role *o = Role::FindByID(id);
  if (o == NULL) {
    CCLOG("%d not found\n", id);
    return;
  }

  // update ship status before shoot
  float x = data["x"].asFloat();
  float y = data["y"].asFloat();
  float angle = data["angle"].asFloat();
  int move = data["move"].asInt();
  int rotate = data["rotate"].asInt();
  int beamid = data["beamid"].asInt();
  // o->set_loc(ccp(x, y));
  // o->set_angle(angle);
  // o->set_move(move, rotate);
  // o->FlushLoc();
  // do shoot
  o->ShootBeam(beamid, x, y, angle);
}

class StopBeamCall : public NetCall {
 public:
  void Proc(Json::Value *value);
};

void StopBeamCall::Proc(Json::Value *value) {
  const Json::Value& v = *value;

  const Json::Value& body = v["body"];
  const Json::Value& data = body["data"];

  int id = data["id"].asInt();
  int beamid = data["beamid"].asInt();

  Role *o;
  if ((o = Role::FindByID(id)) == NULL)
    return;

  o->StopBeam(beamid);
}

class ShipDeadCall : public NetCall {
 public:
  void Proc(Json::Value *value);
};

void ShipDeadCall::Proc(Json::Value *value) {
  const Json::Value& v = *value;

  const Json::Value& body = v["body"];
  const Json::Value& data = body["data"];

  int id = data.asInt();

  Role *o;
  if ((o = Role::FindByID(id)) == NULL)
    return;

  o->Die();
}

class ShipRestartCall : public NetCall {
 public:
  void Proc(Json::Value *value);
};

void ShipRestartCall::Proc(Json::Value *value) {
  const Json::Value& v = *value;
  const Json::Value& body = v["body"];
  const Json::Value& data = body["data"];

  int id = data.asInt();
  Role *o;
  if ((o = Role::FindByID(id)) == NULL)
    return;

  o->Restart();
}

class ShowPathCall : public NetCall {
 public:
  ShowPathCall(CCSprite *sp) { showsp_ = sp; }
  void Proc(Json::Value *value);
  CCSprite *showsp_;
};

void ShowPathCall::Proc(Json::Value *value) {
  const Json::Value& v = *value;

  const Json::Value& body = v["body"];
  const Json::Value& u = body["data"];

  showsp_->setVisible(true);

    float x = u["x"].asFloat();
    float y = u["y"].asFloat();
    float angle = u["angle"].asFloat();

    showsp_->setPosition(ccp(x, y));
    showsp_->setRotation(angle);
}


bool GameLayer::init() {
  if (!CCLayer::init())
    return false;

  setTouchEnabled(true);

  CCLOG("GameLayer init\n");
  return true;
}

void GameLayer::registerWithTouchDispatcher(void) {
  CCDirector::sharedDirector()->getTouchDispatcher()->addStandardDelegate(this, 0);
}

void GameLayer::ccTouchesBegan(CCSet *touches, CCEvent *event) {
}

void GameLayer::ccTouchesMoved(CCSet *touches, CCEvent *event) {
}

void GameLayer::ccTouchesEnded(CCSet *touches, CCEvent *event) {
  CCSetIterator iter = touches->begin();
  for (; iter != touches->end(); iter++) {
    CCTouch* touch = (CCTouch*)(*iter);
    if (Role::Self()->TrySetTarget(touch)) {
      continue;
    }
    CCPoint loc = touch->getLocation();
    Role::Self()->TrySetDest(loc);
  }
}

void GameLayer::ccTouchesCancelled(CCSet *touches, CCEvent *event) {
  CCLOG("%s\n", __FUNCTION__);
}

void GameLayer::InitSelf() {
  CCSize size = CCDirector::sharedDirector()->getWinSize();
  Role::Self()->Init(this, ccp(size.width / 2, size.height / 2));
}

void GameLayer::onEnter() {
  CCLayer::onEnter();
  CCLOG("GameLayer onEnter\n");

  scheduleUpdate();
  // setKeyboardEnabled(true);

  // proc net msg
  addChild(NetNode::Shared());

  npc_ = NPCLayer::create();
  addChild(npc_, 10);
  // npc_->setPosition(ccp(-480, -320));
  npc_->setPosition(ccp(0, 0));

  bg_ = BgLayer::create();
  addChild(bg_, 5);

  // radar_ = CCSprite::create("radio.png");
  // addChild(radar_, 40);
  // radar_->setPosition(ccp(840, 520));

  InitSelf();
  CCProgressTimer *bloodbar = CCProgressTimer::create(CCSprite::create("hp.png"));
  bloodbar->setType(kCCProgressTimerTypeBar);
  bloodbar->setMidpoint(ccp(0, 0));
  bloodbar->setBarChangeRate(ccp(1, 0));
  bloodbar->setAnchorPoint(ccp(0, 0));
  bloodbar->setPosition(ccp(40, 500));
  bloodbar->setPercentage(100);
  Role::Self()->set_bloodbar(bloodbar);
  addChild(bloodbar, 30);

  UserNotifyCall *call = new UserNotifyCall;
  call->npc_ = npc_;

  showsp_ = CCSprite::create("ship-1.png");
  showsp_->setScale(0.5);
  addChild(showsp_, 15);
  showsp_->setVisible(false);

  NetNode::Shared()->AddCallback(3, call);
  NetNode::Shared()->AddCallback(7, new StopBeamCall);
  NetNode::Shared()->AddCallback(8, new ShootBeamCall);
  NetNode::Shared()->AddCallback(9, new ShipDeadCall);
  NetNode::Shared()->AddCallback(10, new ShipRestartCall);
  // NetNode::Shared()->AddCallback(11, new ShowPathCall(showsp_));

  schedule(schedule_selector(GameLayer::TimeCallback), 0.05, -1, 0);
}

void GameLayer::TimeCallback(float dt) {
  // Role::Self()->SendUpdate();
}

void GameLayer::MoveShips(float dt) {
  for (int i = 0; i < 16; i++) {
    Role *r = Role::FindByID(i);
    if (r == NULL || r->sprite() == NULL || r->dead())
      continue;

    // r->UpdateLoc(dt);
    r->UpdateMove(dt);
  }
}

void GameLayer::update(float dt) {
  MoveShips(dt);
}

bool GameScene::init() {
  if (!CCScene::init())
    return false;

  return true;
}

void GameScene::onEnter() {
  CCScene::onEnter();

  GameLayer *layer = GameLayer::create();
  assert(layer != NULL);

  addChild(layer);
}
