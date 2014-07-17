#ifndef __MP_GAME_SCENE_H__
#define __MP_GAME_SCENE_H__

#include <cocos2d.h>
#include <cocos-ext.h>

USING_NS_CC;
USING_NS_CC_EXT;

class BgLayer;
class NPCLayer;

class GameScene : public CCScene {
 public:
  CREATE_FUNC(GameScene);

  bool init();
  void onEnter();
  void update(float dt);

  void TimeCallback(float dt);
  void InitSelf();
  void MoveShips(float dt);
 private:
  BgLayer *bg_;
  NPCLayer *npc_;
  CCSprite *radar_;
};

#endif
