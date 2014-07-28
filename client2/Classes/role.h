#ifndef __MP_ROLE_H__
#define __MP_ROLE_H__

#include <string>
#include "cocos2d.h"

USING_NS_CC;

const int MOVE_NONE = 0;
const int MOVE_FORWARD = 1;
const int MOVE_BACKWARD = 2;

const int ROTATE_NONE = 0;
const int ROTATE_LEFT = 1;
const int ROTATE_RIGHT = 2;

const int MAX_BEAMCOUNT = 5;

const int SCREEN_WIDTH = 960;
const int SCREEN_HEIGHT = 540;

const int MAP_WIDTH = SCREEN_WIDTH;
const int MAP_HEIGHT = SCREEN_HEIGHT;

const int SHIP_SPEED = 160;
const int RADAR_SCALE = 1;

class BeamPool : public CCObject {
 public:
  void RemoveBeam(CCSprite *b) {
    b->stopAllActions();
    b->removeFromParentAndCleanup(true);

    for (int i = 0; i < 5; i++) {
      if (pool_[i] == b) {
        CCLOG("clear ship beam %d\n", i);
        pool_[i] = NULL;
      }
    }
  }

  void RemoveBeam(int id) {
    CCSprite *s = pool_[id];
    if (s == NULL)
      return;

    s->stopAllActions();
    s->removeFromParentAndCleanup(true);
    pool_[id] = NULL;
  }

  void TestBeam(CCNode *n, void *data) {
    // CCLOG("%s %p\n", __func__, n);
    CCSprite *s = (CCSprite *)data;
    RemoveBeam(s);
  }

  BeamPool() {
    memset(pool_, 0, sizeof(pool_));
  }

  int FindID() {
    for (int i = 0; i < 5; i++) {
      if (pool_[i] == NULL)
        return i;
    }
    return -1;
  }

  void AddBeam(int beamid, CCSprite *s) {
    pool_[beamid] = s;
  }

 private:
  CCSprite *pool_[5];
};

class Role {
 public:
  static Role * Create(int id, const std::string& name);
  static Role * CreateSelf(int id, const std::string& name);
  static Role * Create(int id, CCLayer *parent, const CCPoint loc);
  static Role * CreateSelf(int id, CCLayer *parent, const CCPoint loc);
  static Role * Self();
  static Role * FindByID(int id);

  void CalcDestRotate(const CCPoint& dest);
  void CalcDestMove(const CCPoint& dest);
  void CalcDest(const CCPoint& dest) { CalcDestRotate(dest);};
  void SendUpdate();
  void Restart();
  void Die();
  void StopBeam(int beamid);
  void ShootBeam(int beamid, float x, float y, float angle);
  void FlushLoc();
  void UpdateLoc(float dt);
  void Move(float dt);
  void Rotate(float dt);
  void UpdateMove(double dt);
  void Init(CCNode *parent, const CCPoint loc);
  CCSprite * sprite() { return sp_; }
  bool dead() { return dead_; }
  void set_id(int id) { id_ = id; }
  int id() { return id_; }
  void set_angle(float angle) { angle_ = angle; }
  void set_loc(const CCPoint loc) { loc_ = loc; }
  void set_dest(const CCPoint loc) { dest_loc_ = loc; move_ = MOVE_FORWARD; rotate_ = ROTATE_LEFT;}
  void set_rotate_dt(double dt1) { rotate_dt_= dt1; }
  void set_move(int move, int rotate) { move_ = move; rotate_ = rotate; }
  void set_target(Role *target) { target_ = target; }
  void set_bloodbar(CCProgressTimer *b) { bloodbar_ = b; }
  bool ContainTouch(CCTouch *touch);
  bool TrySetTarget(CCTouch *touch);
  void TrySetDest(const CCPoint& dest);
  Role();

 private:
  CCRect rect(void) {
    CCSize s = sp_->getTexture()->getContentSize();
    return CCRectMake(-s.width / 2, -s.height / 2, s.width, s.height);
  }
  void SendSetTarget(int id);
  void doTarget();
  CCSprite *sp_;
  Role *target_;
  CCProgressTimer *bloodbar_;
  CCNode *parent_;
  CCPoint loc_;
  CCPoint dest_loc_;
  double rotate_dt_;
  double move_dt_;
  double angle_;
  int id_;
  int move_;
  int rotate_;
  bool isself_;
  bool dead_;
  int syncstep_;
  std::string name_;
  BeamPool beampool_;
  static Role *self_;
  static Role * table_[16];
};

#endif
