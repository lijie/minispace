package minispace

import "sync"

type Online struct {
	table map[string]*Client
	lock sync.Mutex
}

var onlineTable *Online
func init() {
	onlineTable = NewOnline()
}

func NewOnline() *Online {
	on := &Online{}
	on.table = make(map[string]*Client)
	return on
}

func (on *Online) Insert(name string, c *Client) error {
	on.lock.Lock()
	defer on.lock.Unlock()

	_, ok := on.table[name]
	if ok {
		return ErrUserAlreadyLogin
	}

	on.table[name] = c
	return nil
}

func (on *Online) Search(name string) (c *Client) {
	on.lock.Lock()
	defer on.lock.Unlock()

	c, ok := on.table[name]
	if ok {
		return
	}
	return nil
}

func (on *Online) Delete(name string) {
	on.lock.Lock()
	defer on.lock.Unlock()
	delete(on.table, name)
}

func InsertOnline(name string, c *Client) error {
	return onlineTable.Insert(name, c)
}

func SearchOnline(name string) *Client {
	return onlineTable.Search(name)
}

func DeleteOnline(name string) {
	onlineTable.Delete(name)
}
