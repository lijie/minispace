package minispace

import "testing"

type ListHost struct {
	value int
	list List
}

func TestListPushBack(t *testing.T) {
	host := make([]ListHost, 10)

	for i := 0; i < 10; i++ {
		host[i].value = i
		InitList(&host[i].list, &host[i])
	}

	var head List
	InitList(&head, nil)

	for i := 0; i < 10; i++ {
		head.PushBack(&host[i].list)
	}

	for n := head.Next(); n != &head; n = n.Next() {
		h := n.Host().(*ListHost)
		t.Log("host value", h.value)
	}

	for i := 0; i < 10; i++ {
		host[i].list.RemoveSelf()
	}

	t.Log("pass PushBack")
}
