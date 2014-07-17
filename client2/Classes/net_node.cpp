#include <assert.h>
#include "cocos2d.h"
#include "net_node.h"

USING_NS_CC_EXT;

static const int kMaxCmd = 16;
static NetNode *shared_net_node = NULL;

NetNode * NetNode::Shared() {
  if (shared_net_node == NULL)
    shared_net_node = NetNode::create();
  return shared_net_node;
}

bool NetNode::init() {
    CCNode::init();
    memset(table_, 0, sizeof(table_));
    fifo_ = fifo_new(1024);
    assert(fifo_ != NULL);
    return true;
}

void NetNode::onEnter() {
  CCNode::onEnter();
  scheduleUpdate();
}

void NetNode::onOpen(WebSocket *ws) {
    if (conn_) {
      // TODO: shouldn't call onOpen in net thread,
      // PugMsg() to fifo_ and called in ui thread for safe.
      conn_->onOpen();
    }
}

void NetNode::onClose(WebSocket *ws) {
    if (conn_) {
      // TODO: shouldn't call onOpen in net thread,
      // PugMsg() to fifo_ and called in ui thread for safe.
      conn_->onClose();
    }
}

void NetNode::onMessage(WebSocket *ws, const WebSocket::Data& data) {
  // printf("%s\n", __func__);

  Json::Reader reader;
  Json::Value *result = new Json::Value;
  assert(result != NULL);
  if (!reader.parse(std::string(data.bytes), *result, false)) {
    CCLOG("parse %s error", data.bytes);
    return;
  }

  PutMsg(result);
}

void NetNode::onError(WebSocket* ws, const WebSocket::ErrorCode& error) {
    if (conn_) {
      conn_->onClose();
    }
}

bool NetNode::Connect(const char *url, NetConn *conn) {
    ws_ = new WebSocket;
    ws_->init(*this, url);
    conn_ = conn;
    return true;
}

bool NetNode::Send(const Json::Value& value) {
  Json::FastWriter writer;
  std::string content = writer.write(value);
  ws_->send(content);
  return true;
}

void NetNode::PutMsg(Json::Value *value) {
    uintptr_t d = (uintptr_t)value;
    if (!fifo_full(fifo_)) {
        fifo_put(fifo_, d);
    } else {
      delete value;
    }
}

Json::Value * NetNode::GetMsg() {
  // CCLOG("%s\n", __func__);
  if (!fifo_empty(fifo_)) {
    return (Json::Value *)fifo_get(fifo_);
  }
  return NULL;
}

void NetNode::AddCallback(int cmd, NetCall *call) {
  if (cmd < 0 || cmd >= kMaxCmd)
    return;

  table_[cmd] = call;
}

void NetNode::update(float dt) {
  Json::Value *v = GetMsg();
  if (v == NULL)
    return;

  int cmd = v->get("cmd", kMaxCmd).asInt();
  if (cmd >= kMaxCmd)
    return;

  NetCall *call = table_[cmd];
  if (call) {
    // call cmd callback
    // CCLOG("cmd %d call\n", cmd);
    call->Proc(v);
  }

  // done
  delete v;
}

class NetNodeTestOpen : public NetCall {
  public:
    NetNode *node;
    void Proc(Json::Value *value) {
        Json::Value foo;
        node->Send(foo);
    }
};

#if 0
void NetNodeTest() {
    NetNode *node = new NetNode;
    NetNodeTestOpen open;
    open.node = node;

    node->init();
    node->Connect("ws://127.0.0.1:12345/echo");
}
#endif
