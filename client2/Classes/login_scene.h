#ifndef __MP_LOGIN_SCENE_H__
#define __MP_LOGIN_SCENE_H__

#include <cocos2d.h>
#include <cocos-ext.h>

USING_NS_CC;
USING_NS_CC_EXT;

using namespace ui;

class LoginScene : public CCScene {
 public:
  virtual bool init();
  virtual void onEnter();
  void update(float dt);
  CREATE_FUNC(LoginScene);

  void onBtnLogin(CCObject* sender, TouchEventType type);
  void startConnect();
  void startLogin();
  void startPlay();

  void InputNameEvent(CCObject *pSender, TextFiledEventType type);
 private:
  int state_;
  UILayer *ui_layer_;
  TextField *pass_;
  TextField *name_;
};


#endif
