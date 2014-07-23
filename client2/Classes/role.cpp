#include <math.h>
#include "role.h"
#include "net_node.h"
#include "json/json.h"

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
  move_dt_ = 0;
  rotate_dt_ = 0;
  target_ = NULL;
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
  double angle;

  if (rotate_ == ROTATE_NONE)
    return;

  CalcDestRotate(dest_loc_);
  if (rotate_dt_ <= 0) {
    rotate_ = ROTATE_NONE;
    return;
  }

  if (rotate_dt_ < dt) {
    dt = rotate_dt_;
    rotate_dt_ = 0;
  } else {
    rotate_dt_ -= dt;
  }

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

  if (rotate_dt_ == 0)
    rotate_ = ROTATE_NONE;
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

void Role::CalcDestMove(const CCPoint& dest) {
  CCPoint loc2 = sp_->getPosition();
  double y = dest.y - loc2.y;
  double x = dest.x - loc2.x;
  double dist = sqrt(y*y + x*x);
  move_dt_ = dist / 160; 
}

void Role::CalcDestRotate(const CCPoint& dest) {
  CCPoint loc2 = sp_->getPosition();
  double r = atan((dest.y - loc2.y) / (dest.x - loc2.x));

    r = r / M_PI * 180;
    double r2 = sp_->getRotation();

    double r3 = 0;
    if (r < 0) {
      if (dest.y > loc2.y)
        r3 = 360 - (180 + r);
      else
        r3 = abs(r);
    } else {
      if (dest.x > loc2.x)
        r3 = 360 - r;
      else
        r3 = 180 - r;
    }

    double r4 = r3 - r2;
    if (r4 < 0)
      r4 = r4 + 360;

    // CCLOG("rrr %f %f %f %f", r, r2, r3, r4);
#if 1
    // move_ = MOVE_FORWARD;
    if (r4 < 180) {
      rotate_ = ROTATE_RIGHT;
      rotate_dt_ = r4 / 120;
    } else {
      rotate_ = ROTATE_LEFT;
      rotate_dt_ = (360 - r4) / 120;
    }
#endif
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
  body["destx"] = dest_loc_.x;
  body["desty"] = dest_loc_.y;
  body["angle"] = sp_->getRotation();
  body["rotate"] = rotate_;
  body["move"] = move_;

  obj["cmd"] = 2;
  obj["userid"] = "test";
  obj["body"] = body;

  NetNode::Shared()->Send(obj);
}

void Role::ShootBeam(int beamid) {
#if 0
  if (isself_) {
    beamid = beampool_.FindID();
    if (beamid == -1)
      return;
  }
#endif
  CCLOG("shot beam %d\n", beamid);

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
  // CCLOG("stop ship %d beam %d\n", id_, beamid);
  beampool_.RemoveBeam(beamid);
}

void Role::Move(float dt) {
  if (move_ == MOVE_NONE)
    return;

  CalcDestMove(dest_loc_);
  if (move_dt_ <= 0) {
    move_ = MOVE_NONE;
    sp_->setPosition(dest_loc_);
    return;
  }

  move_dt_ -= dt;
  if (move_dt_ < 0) {
    move_ = MOVE_NONE;
    move_dt_ = 0;
    sp_->setPosition(dest_loc_);
    return;
  }

  float angle = sp_->getRotation() + 90;

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

bool Role::ContainTouch(CCTouch *touch) {
  return rect().containsPoint(sp_->convertTouchToNodeSpaceAR(touch));
}

void Role::SendSetTarget(int id) {
  Json::Value obj;
  Json::Value body;

  body["x"] = sp_->getPositionX();
  body["y"] = sp_->getPositionY();
  body["angle"] = sp_->getRotation();
  body["rotate"] = rotate_;
  body["move"] = move_;
  body["targetid"] = id;

  obj["cmd"] = 12;
  obj["userid"] = "test";
  obj["body"] = body;

  NetNode::Shared()->Send(obj);
}

bool Role::TrySetTarget(CCTouch *touch) {
  for (int i = 0; i < 16; i++) {
    Role *r = Role::FindByID(i);
    if (r == NULL || r->sprite() == NULL || r->dead())
      continue;

    if (r == this)
      continue;

    if (r->ContainTouch(touch)) {
      CCLOG("set target %d", r->id());
      set_target(r);
      SendSetTarget(r->id());
      return true;
    }
  }

  return false;
}

void Role::TrySetDest(const CCPoint& dest) {
  target_ = NULL;
  set_dest(dest);
  SendUpdate();
}

void Role::doTarget() {
  if (target_ == NULL)
     return;
  if (target_->dead()) {
    target_ = NULL;
    return;
  }

  CCPoint dest = target_->sprite()->getPosition();
  CCPoint loc2 = sp_->getPosition();
  double y = dest.y - loc2.y;
  double x = dest.x - loc2.x;
  double dist = sqrt(y*y + x*x);
  set_dest(target_->sprite()->getPosition());
  if (dist < 100) { 
    move_ = MOVE_NONE;
  }
}

void Role::UpdateMove(double dt) {
  doTarget();
  Rotate(dt);
  Move(dt);
}
