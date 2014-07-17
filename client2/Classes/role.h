#ifndef __MP_ROLE_H__
#define __MP_ROLE_H__

#include <string>
#include "cocos2d.h"

USING_NS_CC;

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

  void SendUpdate();
  void Restart();
  void Die();
  void StopBeam(int beamid);
  void ShootBeam(int beamid);
  void FlushLoc();
  void UpdateLoc(float dt);
  void Move(float dt);
  void Rotate(float dt);
  void Init(CCNode *parent, const CCPoint loc);
  CCSprite * sprite() { return sp_; }
  bool dead() { return dead_; }
  void set_id(int id) { id_ = id; }
  int id() { return id_; }
  void set_angle(float angle) { angle_ = angle; }
  void set_loc(const CCPoint loc) { loc_ = loc; }
  void set_move(int move, int rotate) { move_ = move; rotate_ = rotate; }
  Role();

 private:
  CCSprite *sp_;
  CCNode *parent_;
  CCPoint loc_;
  float angle_;
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
