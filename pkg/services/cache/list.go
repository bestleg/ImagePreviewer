package cache

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
	Value      interface{}
	Next, Prev *ListItem
}

type list struct {
	front, back *ListItem
	len         int
}

func NewList() List {
	return &list{}
}

func (l list) Len() int {
	return l.len
}

func (l list) Front() *ListItem {
	return l.front
}

func (l list) Back() *ListItem {
	return l.back
}

func (l *list) PushFront(v interface{}) *ListItem {
	newFirstItem := &ListItem{Value: v}

	if l.back == nil {
		l.back = newFirstItem
	} else {
		newFirstItem.Next = l.front
		l.front.Prev = newFirstItem
	}

	l.front = newFirstItem
	l.len++

	return newFirstItem
}

func (l *list) PushBack(v interface{}) *ListItem {
	i := &ListItem{Value: v}

	if l.front == nil {
		l.front = i
	} else {
		i.Prev = l.back
		l.back.Next = i
	}

	l.back = i
	l.len++

	return i
}

func (l *list) Remove(i *ListItem) {
	if i.Next != nil {
		i.Next.Prev = i.Prev
	} else {
		l.front = i.Prev
	}

	if i.Prev != nil {
		i.Prev.Next = i.Next
	} else {
		l.back = i.Next
	}

	l.len--
}

func (l *list) MoveToFront(i *ListItem) {
	// l.Remove(i)
	// l.PushFront(i.Value)
	if i.Prev != nil {
		i.Prev.Next = i.Next
	} else {
		return
	}
	if i.Next != nil {
		i.Next.Prev = i.Prev
	}
	i.Prev = nil
	i.Next = l.front
	l.front = i
}
