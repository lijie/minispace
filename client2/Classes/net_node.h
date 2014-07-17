#ifndef __MP_NET_NODE_H__
#define __MP_NET_NODE_H__

#include "network/WebSocket.h"
#include "fifo.h"
#include "json/json.h"

USING_NS_CC_EXT;

class NetCall {
  public:
    virtual void Proc(Json::Value *value) = 0;
};

class NetConn {
 public:
  virtual void onOpen() = 0;
  virtual void onClose() = 0;
};

class NetNode : public cocos2d::CCNode, public WebSocket::Delegate {
 public:
  bool init();
  void update(float dt);
  void onEnter();
  CREATE_FUNC(NetNode);

  bool Connect(const char *url, NetConn *);
  bool Send(const Json::Value& value);
  void AddCallback(int cmd, NetCall *call);

  // called in net thread
  void PutMsg(Json::Value *v);
  // called in ui thread
  Json::Value * GetMsg();

  // for websocket
  virtual void onOpen(WebSocket *ws);
  virtual void onMessage(WebSocket* ws, const WebSocket::Data& data);
  virtual void onClose(WebSocket* ws);
  virtual void onError(WebSocket* ws, const WebSocket::ErrorCode& error);

  static NetNode * Shared();
 private:
  WebSocket *ws_;
  fifo_t *fifo_;
  NetConn *conn_;
  NetCall *table_[16];
};

#endif
