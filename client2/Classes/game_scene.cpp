#include <cocos2d.h>
// #include <cocos-ext.h>

class GameScene : public CCScene {
 public:
  CREATE_FUNC(GameScene);

  bool init();
  void onEnter();

 private:
};

bool GameScene::init() {
  if (!CCScene::init())
    return false;

  return true;
}

void GameScene::onEnter() {
}

