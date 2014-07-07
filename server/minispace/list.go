package minispace

type List struct {
	next *List
	prev *List
	host interface{}
}

func InitList(l *List, host interface{}) {
	l.next = l
	l.prev = l
	l.host = host
}

func add(n, prev, next *List) {
	next.prev = n
	n.next = next
	n.prev = prev
	prev.next = n
}

func del(prev, next *List) {
	next.prev = prev
	prev.next = next
}

func (l *List) PushBack(n *List) {
	add(n, l.prev, l)
}

func (l *List) PushFront(n *List) {
	add(n, l, l.next)
}

func (l *List) Next() *List {
	return l.next
}

func (l *List) Prev() *List {
	return l.prev
}

func (l *List) Host() interface{} {
	return l.host
}

func (l *List) RemoveSelf() {
	del(l.prev, l.next)
	InitList(l, l.host)
}
