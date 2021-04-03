package tc

import (
	"fmt"
	"sync"
)

/*
 * General queue interface
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

//node base info
type Node struct {
	data interface{}
	next *Node
}

//queue info
type Queue struct {
	head *Node
	tail *Node
	size int
	sync.Mutex
}

//construct
func NewQueue() *Queue {
	//self init
	this := &Queue{
		head:nil,
		tail:nil,
		size:0,
	}
	return this
}

//check empty
func (q *Queue) IsEmpty() bool {
	return q.size == 0
}

//get size
func (q *Queue) GetSize() int {
	return q.size
}

//check node is exist or not
func (q *Queue) CheckNode(node *Node) bool {
	p := q.head
	for p != nil {
		if p == node {
			return true
		}
		p = p.next
	}
	return false
}

//loop
func (q *Queue) Traverse() {
	if q.IsEmpty() {
		return
	}
	p := q.head
	for p != nil {
		fmt.Println(p.data, " ")
		p = p.next
	}
}

//delete node of data
func (q *Queue) DelNode(data interface{}) bool {
	var (
		prev *Node
		cur *Node
	)

	cur = q.head
	if cur == nil {
		return false
	}
	q.Lock()
	for cur != nil {
		if cur.data == data {
			//let current point to prev node
			if prev != nil {
				cur = prev
			}
			//let next point to next.next node
			cur.next = cur.next.next
		}else{
			prev = cur
			cur = cur.next
		}
	}
	q.Unlock()
	return true
}

//get node of assigned data
func (q *Queue) GetNode(data interface{}) *Node {
	p := q.head
	for p != nil {
		if p.data == data {
			//find data
			return p
		}else{
			p = p.next
		}
	}
	return nil
}

//pop element
func (q *Queue) Pop() interface{} {
	if q.head == nil {
		return nil
	}

	q.Lock()
	data := q.head.data
	q.head = q.head.next
	q.size--
	q.Unlock()

	return data
}

//push new element
func (q *Queue) Push(data interface{}) {
	//init node
	node := Node{
		data:data,
		next:nil,
	}
	q.Lock()
	if q.size == 0 {
		q.head = &node
		q.tail = &node
	}else{
		q.tail.next = &node
		q.tail = &node
	}
	q.size++
	q.Unlock()
}


