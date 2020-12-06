package calgorithm

type TCircleQueue struct {
	datas []interface{}
	head  int
	lenth int
}

func NewCircleQueue(size int) *TCircleQueue {
	if 0 >= size {
		return nil
	}

	return &TCircleQueue{
		datas: make([]interface{}, size),
	}
}

func (this *TCircleQueue) Push(data interface{}) {
	idx := (this.head + this.lenth) % len(this.datas)
	this.datas[idx] = data
	this.lenth++
	if this.lenth > len(this.datas) {
		this.lenth = len(this.datas)
		this.head = this.nextIdx(this.head)
	}
}

func (this *TCircleQueue) Pop() interface{} {
	if this.lenth <= 0 {
		return nil
	}
	this.lenth--
	ret := this.datas[this.head]
	this.head = this.nextIdx(this.head)
	return ret
}

func (this *TCircleQueue) Length() int {
	return this.lenth
}

func (this *TCircleQueue) Item(idx int) interface{} {
	idx = (idx + this.head) % len(this.datas)
	return this.datas[idx]
}

func (this *TCircleQueue) nextIdx(idx int) int {
	return (idx + 1) % len(this.datas)
}
