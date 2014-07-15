#include "cocos2d.h"
#include "net_node.h"

USING_NS_CC_EXT;

static const int kMaxCmd = 16;
static NetNode shared_net_node;

NetNode * NetNode::Shared() {
  return &shared_net_node;
}

bool NetNode::init() {
    CCNode::init();
    return true;
}

void NetNode::onOpen(WebSocket *ws) {
    if (conn_) {
      conn_->onOpen();
    }
}

void NetNode::onClose(WebSocket *ws) {
    if (conn_) {
      conn_->onClose();
    }
}

void NetNode::onMessage(WebSocket *ws, const WebSocket::Data& data) {
    printf("%s\n", __func__);
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
    ws_->send("hello, world");
    return true;
}

void NetNode::PutMsg(Json::Value *value) {
    uintptr_t d = (uintptr_t)value;
    if (!fifo_full(fifo_)) {
        fifo_put(fifo_, d);
    }
}

Json::Value * NetNode::GetMsg() {
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
    // call cmd callback
    call->Proc(v);

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
