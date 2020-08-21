package main

// Queue 队列信息
type Queue struct {
	list    *SingleList
	isInUse bool
}

// Init 队列初始化
func (q *Queue) Init() {
	q.list = new(SingleList)
	q.list.Init()
	q.isInUse = false
}

// Size 获取队列长度
func (q *Queue) Size() uint {
	return q.list.Size
}

// Enqueue 进入队列
func (q *Queue) Enqueue(data interface{}) bool {
	q.isInUse = true
	return q.list.Append(&SingleNode{Data: data})
}

// Dequeue 出列
func (q *Queue) Dequeue() interface{} {
	q.isInUse = true
	node := q.list.Get(0)
	if node == nil {
		return nil
	}
	q.list.Delete(0)
	return node.Data
}

// Peek 查看队头信息
func (q *Queue) Peek() interface{} {
	node := q.list.Get(0)
	if node == nil {
		return nil
	}
	return node.Data
}

// Contains 队列中包含值
func (q *Queue) Contains(value interface{}) bool {
	for i := uint(0); i < q.Size(); i++ {
		gdata := q.list.Get(i)
		if gdata != nil {
			if value == gdata.Data {
				return true
			}
		} else {
			return false
		}
	}
	return false
}

// 获取队列中某个值
func (q *Queue) GetData(index uint) interface{} {
	if index >= q.Size() {
		return nil
	}
	node := q.list.Get(index)
	if node == nil {
		return nil
	}
	return node.Data
}

func (q *Queue) ChangeStatus(b bool) {
	q.isInUse = b
}
func (q *Queue) GetStatus() bool {
	return q.isInUse
}
