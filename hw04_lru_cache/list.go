package hw04lrucache

type List interface {
	Len() int
	Front() *ListItem
	Back() *ListItem
	PushFront(v interface{}) *ListItem
	PushBack(v interface{}) *ListItem
	Remove(i *ListItem)
	MoveToFront(i *ListItem)
}

type ListItem struct {
	Value interface{}
	Next  *ListItem
	Prev  *ListItem
}

type list struct {
	head   *ListItem
	tail   *ListItem
	length int
}

func NewList() List {
	return &list{}
}

func (l *list) Len() int {
	return l.length
}

func (l *list) Front() *ListItem {
	return l.head
}

func (l *list) Back() *ListItem {
	return l.tail
}

func (l *list) PushFront(v interface{}) *ListItem {
	item := &ListItem{Value: v}
	if l.length == 0 {
		l.head = item
		l.tail = item
	} else {
		item.Next = l.head
		l.head.Prev = item
		l.head = item
	}
	l.length++
	return item
}

func (l *list) PushBack(v interface{}) *ListItem {
	item := &ListItem{Value: v}
	if l.length == 0 {
		l.head = item
		l.tail = item
	} else {
		item.Prev = l.tail
		l.tail.Next = item
		l.tail = item
	}
	l.length++
	return item
}

func (l *list) Remove(i *ListItem) {
	if i.Prev != nil {
		i.Prev.Next = i.Next
	} else {
		l.head = i.Next
	}

	if i.Next != nil {
		i.Next.Prev = i.Prev
	} else {
		l.tail = i.Prev
	}

	i.Next = nil
	i.Prev = nil

	l.length--
}

func (l *list) MoveToFront(i *ListItem) {
	if i == l.head {
		return
	}
	if i.Prev != nil {
		i.Prev.Next = i.Next
	} else {
		l.head = i.Next
	}
	if i.Next != nil {
		i.Next.Prev = i.Prev
	} else {
		l.tail = i.Prev
	}
	i.Next = l.head
	i.Prev = nil
	if l.head != nil {
		l.head.Prev = i
	}
	l.head = i
	if l.tail == nil {
		l.tail = i
	}
}
