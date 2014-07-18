#ifndef __MP_GAME_SCENE_H__
#define __MP_GAME_SCENE_H__

#include <cocos2d.h>
#include <cocos-ext.h>

USING_NS_CC;
USING_NS_CC_EXT;

class BgLayer;
class NPCLayer;

class GameLayer : public CCLayer {
 public:
  CREATE_FUNC(GameLayer);

  bool init();
  void onEnter();
  void update(float dt);

  void TimeCallback(float dt);
  void InitSelf();
  void MoveShips(float dt);

  // for touche
  void registerWithTouchDispatcher(void);
  void ccTouchesBegan(CCSet *touches, CCEvent *event);
  void ccTouchesMoved(CCSet *touches, CCEvent *event);
  void ccTouchesEnded(CCSet *touches, CCEvent *event);
  void ccTouchesCancelled(CCSet *touches, CCEvent *event);

 private:
  BgLayer *bg_;
  NPCLayer *npc_;
  CCSprite *radar_;
};

class GameScene : public CCScene {
 public:
  CREATE_FUNC(GameScene);
  bool init();
  void onEnter();
};

#endif
