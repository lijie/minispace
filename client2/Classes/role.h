#ifndef __MP_ROLE_H__
#define __MP_ROLE_H__

class Role {
 public:
  static Role * Create(int id);

 private:
  CCSprite *sp_;
  CCLayer *parent_;
  CCPoint loc_;
  float angle_;
  int id_;
  int move_;
  int rotate_;
  bool self_;
  string name_;
};

Role * Role::Create(int id) {
  role = new Role();

  if (role == NULL)
    return NULL;

  role->id_ = id;
  return role;
}

#endif
