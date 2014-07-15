#include "cocos2d.h"
#include "network/WebSocket.h"
#include "fifo.h"
#include "json/json.h"

USING_NS_CC_EXT;

class NetCall {
  public:
    int cmd;
    virtual void Proc(Json::Value *value) = 0;
};

class NetNode;

class NetSocket : public WebSocket::Delegate {
  public:
    bool init(const char *url);
    ~NetSocket();
    virtual void onOpen(WebSocket *ws);
    virtual void onMessage(WebSocket* ws, const Data& data);
    virtual void onClose(WebSocket* ws);
    virtual void onError(WebSocket* ws, const ErrorCode& error);

  private:
    WebSocket *ws_;
    NetNode *node_;
};

NetSocket::init(const char *url) {
    ws_ = new WebSocket;
    ws_->init(*this, url);
}

NetSocket::~NetSocket() {
    if (ws_) {
        delete ws_;
    }
}

static const int kMaxCmd = 16;

class NetNode : public cocos2d::CCNode {
  public:
    bool init();
    void update(float dt);

    void AddCallback(int cmd);

    // called in net thread
    void PutMsg(Json::Value *v);
    // called in ui thread
    Json::Value * GetMsg();

  private:
    NetSocket sock_;
    fifo_t *fifo_;
    NetCall *table_[16];
};

bool NetNode::init(const char *url) {
    CCNode::init();
    sock_.init(url);
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
}
