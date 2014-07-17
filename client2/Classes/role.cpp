#include <math.h>
#include "role.h"
#include "net_node.h"
#include "json/json.h"

const int MOVE_NONE = 0;
const int MOVE_FORWARD = 1;
const int MOVE_BACKWARD = 2;

const int ROTATE_NONE = 0;
const int ROTATE_LEFT = 1;
const int ROTATE_RIGHT = 2;

const int MAX_BEAMCOUNT = 5;

const int SCREEN_WIDTH = 960;
const int SCREEN_HEIGHT = 540;

const int MAP_WIDTH = SCREEN_WIDTH * 2;
const int MAP_HEIGHT = SCREEN_HEIGHT * 2;

const int SHIP_SPEED = 160;
const int RADAR_SCALE = 1;

Role * Role::self_;
Role * Role::table_[16] = {NULL};

void RoleModuleInit() {
}

Role::Role() {
  sp_ = NULL;
  parent_ = NULL;
  id_ = -1;
  isself_ = false;
  dead_ = false;
  move_ = MOVE_NONE;
  rotate_ = ROTATE_NONE;
  syncstep_ = 0;
}

Role * Role::FindByID(int id) {
  for (int i = 0; i < 16; i++) {
    if (table_[i] == NULL)
      continue;

    if (table_[i]->id_ == id)
      return table_[i];
  }

  return NULL;
}

Role * Role::Create(int id, const std::string& name) {
  Role *r = new Role;
  r->id_ = id;
  r->name_ = name;
  table_[id] = r;
  return r;
}

Role * Role::CreateSelf(int id, const std::string& name) {
  Role *r = Create(id, name);
  if (r) {
    r->isself_ = true;
  }
  self_ = r;
  return r;
}

Role * Role::Create(int id, CCLayer *parent, const CCPoint loc) {
  Role *r = new Role;
  r->id_ = id;
  r->Init(parent, loc);
  table_[id] = r;
  return r;
}

Role * Role::CreateSelf(int id, CCLayer *parent, const CCPoint loc) {
  Role *r = Create(id, parent, loc);
  if (r) {
    r->isself_ = true;
  }
  self_ = r;
  return r;
}

void Role::Init(CCNode *parent, const CCPoint loc) {
  if (sp_ != NULL)
    return;

  char buf[32];
  snprintf(buf, 32, "ship-%d.png", id_);
  sp_ = CCSprite::create(buf);
  assert(sp_ != NULL);

  sp_->setAnchorPoint(ccp(0.5, 0.5));
  sp_->setPosition(loc);
  sp_->setScale(0.5);
  parent->addChild(sp_, 20);
  parent_ = parent;
}

Role * Role::Self() {
  return Role::self_;
}

void Role::Restart() {
  if (isself_) {
    loc_ = ccp(SCREEN_WIDTH / 2, SCREEN_HEIGHT / 2);
  }

  angle_ = 0;
  move_ = MOVE_NONE;
  rotate_ = ROTATE_NONE;
  dead_ = false;

  sp_->setPosition(loc_);
  sp_->setRotation(0);
  sp_->setVisible(true);
  if (isself_) {
    // TODO
    // bloodbar_->setPercentage(100);
  }
}

void Role::Die() {
  dead_ = true;
  sp_->setVisible(false);

  if (isself_) {
    // TODO:
    // bloodbar_->setPercentage(0);
  }

  CCLOG("ship %d dead\n", id_);
}

void Role::Rotate(float dt) {
  float angle;

  if (rotate_ == ROTATE_LEFT)
    angle = sp_->getRotation() - (120 * dt);
  else if (rotate_ == ROTATE_RIGHT)
    angle = sp_->getRotation() + (120 * dt);
  else
    return;

  if (angle < 0)
    angle = angle + 360;
  else if (angle >= 360)
    angle = angle - 360;
  sp_->setRotation(angle);
}

void Role::UpdateLoc(float dt) {
  if (isself_)
    return;
  syncstep_++;
  if (syncstep_ == 10) {
    FlushLoc();
    syncstep_ = 0;
  }
}

void Role::FlushLoc() {
  sp_->setPosition(loc_);
  sp_->setRotation(angle_);
}

void Role::SendUpdate() {
  Json::Value obj;
  Json::Value body;

  body["x"] = sp_->getPositionX();
  body["y"] = sp_->getPositionY();
  body["angle"] = sp_->getRotation();
  body["rotate"] = rotate_;
  body["move"] = move_;

  obj["cmd"] = 2;
  obj["userid"] = "test";
  obj["body"] = body;

  NetNode::Shared()->Send(obj);
}

void Role::ShootBeam(int beamid) {
  if (isself_) {
    beamid = beampool_.FindID();
    if (beamid == -1)
      return;
  }

  CCSprite *_beam = CCSprite::create("beam-1x.png");
  _beam->setPosition(sp_->getPosition());
  parent_->addChild(_beam, 25);

  float angle = sp_->getRotation() + 90;
  _beam->setRotation(sp_->getRotation());

  float x = 1000 * sin(angle / 180 * M_PI);
  float y = 1000 * cos(angle / 180 * M_PI);

  CCSequence *action = CCSequence::create(
      CCMoveBy::create(3.0, ccp(x, y)),
      CCCallFuncND::create((CCObject *)&beampool_, callfuncND_selector(BeamPool::TestBeam), _beam),
      // CCRemoveSelf::create(true),
      NULL);
  _beam->runAction(action);
  beampool_.AddBeam(beamid, _beam);
  if (isself_) {
    // need report to server
  }
}

void Role::StopBeam(int beamid) {
  CCLOG("stop ship %d beam %d\n", id_, beamid);
  beampool_.RemoveBeam(beamid);
}

void Role::Move(float dt) {
  if (move_ == MOVE_NONE)
    return;

  float angle = sp_->getRotation() + 90;
//  if (move_ == MOVE_BACKWARD)
//    angle -= 180;
//  if (angle >= 360)
//    angle = angle - 360;

  float r = SHIP_SPEED * dt;
  float x = r * sin(angle / 180 * M_PI);
  float y = r * cos(angle / 180 * M_PI);

  if (move_ == MOVE_FORWARD) {
    x = sp_->getPositionX() + x;
    y = sp_->getPositionY() + y;
  } else {
    x = sp_->getPositionX() - x;
    y = sp_->getPositionY() - y;
  }

  if (x > MAP_WIDTH)
    x = MAP_WIDTH;
  else if (x < 0)
    x = 0;

  if (y > MAP_HEIGHT)
    y = MAP_HEIGHT;
  else if (y < 0)
    y = 0;

  sp_->setPosition(ccp(x, y));
  // this.moveRadar(x, y);
}
