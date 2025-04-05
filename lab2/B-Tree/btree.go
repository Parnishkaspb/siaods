package B_Tree

import (
	"fmt"
	"golang.org/x/exp/constraints"
	"strings"
)

type BTreeNode[T constraints.Ordered] struct {
	Leaf     bool
	Keys     []T
	Children []*BTreeNode[T]
}

type BTree[T constraints.Ordered] struct {
	Root *BTreeNode[T]
	t    int // минимальный порядок (t >= 2)
}

func NewBTree[T constraints.Ordered](t int) *BTree[T] {
	if t < 2 {
		panic("Минимальный порядок B-Tree должен быть >= 2")
	}
	return &BTree[T]{Root: &BTreeNode[T]{Leaf: true}, t: t}
}

// ---------- ВСТАВКА ----------

func (t *BTree[T]) Insert(k T) {
	r := t.Root
	if len(r.Keys) == 2*t.t-1 {
		s := &BTreeNode[T]{Leaf: false, Children: []*BTreeNode[T]{r}}
		t.Root = s
		t.splitChild(s, 0)
		t.insertNonFull(s, k)
	} else {
		t.insertNonFull(r, k)
	}
}

func (t *BTree[T]) insertNonFull(x *BTreeNode[T], k T) {
	i := len(x.Keys) - 1
	if x.Leaf {
		x.Keys = append(x.Keys, k)
		for i >= 0 && k < x.Keys[i] {
			x.Keys[i+1] = x.Keys[i]
			i--
		}
		x.Keys[i+1] = k
	} else {
		for i >= 0 && k < x.Keys[i] {
			i--
		}
		i++
		if len(x.Children[i].Keys) == 2*t.t-1 {
			t.splitChild(x, i)
			if k > x.Keys[i] {
				i++
			}
		}
		t.insertNonFull(x.Children[i], k)
	}
}

func (t *BTree[T]) splitChild(x *BTreeNode[T], i int) {
	y := x.Children[i]
	z := &BTreeNode[T]{Leaf: y.Leaf}
	mid := t.t - 1

	if len(y.Keys) <= mid {
		panic(fmt.Sprintf("Cannot split: y.Keys too short (%d <= %d)", len(y.Keys), mid))
	}

	// Сохраняем midKey ДО обрезки y.Keys
	midKey := y.Keys[mid]

	// делим ключи
	z.Keys = append(z.Keys, y.Keys[mid+1:]...)
	y.Keys = y.Keys[:mid]

	// делим детей, если не лист
	if !y.Leaf {
		z.Children = append(z.Children, y.Children[mid+1:]...)
		y.Children = y.Children[:mid+1]
	}

	// вставляем midKey в родителя
	x.Keys = append(x.Keys[:i], append([]T{midKey}, x.Keys[i:]...)...)
	x.Children = append(x.Children[:i+1], append([]*BTreeNode[T]{z}, x.Children[i+1:]...)...)
}

// ---------- ПОИСК ----------

func (n *BTreeNode[T]) Search(k T) (*BTreeNode[T], int) {
	i := 0
	for i < len(n.Keys) && k > n.Keys[i] {
		i++
	}
	if i < len(n.Keys) && k == n.Keys[i] {
		return n, i
	}
	if n.Leaf {
		return nil, -1
	}
	return n.Children[i].Search(k)
}

// ---------- УДАЛЕНИЕ ----------

func (t *BTree[T]) Delete(k T) {
	if t.Root == nil {
		return
	}
	t.Root.delete(k, t.t)
	if len(t.Root.Keys) == 0 && !t.Root.Leaf {
		t.Root = t.Root.Children[0]
	}
}

func (n *BTreeNode[T]) delete(k T, t int) {
	idx := n.findKey(k)

	if idx < len(n.Keys) && n.Keys[idx] == k {
		if n.Leaf {
			n.Keys = append(n.Keys[:idx], n.Keys[idx+1:]...)
		} else {
			if idx+1 < len(n.Children) && len(n.Children[idx+1].Keys) >= t {
				succ := n.Children[idx+1].getSuccessor()
				n.Keys[idx] = succ
				n.Children[idx+1].delete(succ, t)
			} else if idx > 0 && len(n.Children[idx-1].Keys) >= t {
				pred := n.Children[idx-1].getPredecessor()
				n.Keys[idx] = pred
				n.Children[idx-1].delete(pred, t)
			} else {
				var mergedIdx int
				if idx < len(n.Children)-1 {
					n.merge(idx)
					mergedIdx = idx
				} else {
					mergedIdx = idx - 1
					n.merge(mergedIdx)
				}

				// safety check
				if mergedIdx >= 0 && mergedIdx < len(n.Children) {
					n.Children[mergedIdx].delete(k, t)
				}
			}
		}
	} else {
		if n.Leaf {
			return
		}
		if idx >= len(n.Children) {
			return // safety check
		}
		if len(n.Children[idx].Keys) < t {
			n.fill(idx, t)
		}
		if idx < len(n.Children) {
			n.Children[idx].delete(k, t)
		}
	}
}

func (n *BTreeNode[T]) findKey(k T) int {
	idx := 0
	for idx < len(n.Keys) && n.Keys[idx] < k {
		idx++
	}
	return idx
}

func (n *BTreeNode[T]) getPredecessor() T {
	cur := n
	for !cur.Leaf {
		cur = cur.Children[len(cur.Children)-1]
	}
	return cur.Keys[len(cur.Keys)-1]
}

func (n *BTreeNode[T]) getSuccessor() T {
	cur := n
	for !cur.Leaf {
		cur = cur.Children[0]
	}
	return cur.Keys[0]
}

func (n *BTreeNode[T]) merge(idx int) {
	child := n.Children[idx]
	sibling := n.Children[idx+1]

	child.Keys = append(child.Keys, n.Keys[idx])
	child.Keys = append(child.Keys, sibling.Keys...)

	if !child.Leaf {
		child.Children = append(child.Children, sibling.Children...)
	}

	n.Keys = append(n.Keys[:idx], n.Keys[idx+1:]...)
	n.Children = append(n.Children[:idx+1], n.Children[idx+2:]...)
}

func (n *BTreeNode[T]) fill(idx int, t int) {
	if idx != 0 && len(n.Children[idx-1].Keys) >= t {
		n.borrowFromPrev(idx)
	} else if idx != len(n.Children)-1 && len(n.Children[idx+1].Keys) >= t {
		n.borrowFromNext(idx)
	} else {
		if idx != len(n.Children)-1 {
			n.merge(idx)
		} else {
			n.merge(idx - 1)
		}
	}
}

func (n *BTreeNode[T]) borrowFromPrev(idx int) {
	child := n.Children[idx]
	sibling := n.Children[idx-1]

	child.Keys = append([]T{n.Keys[idx-1]}, child.Keys...)
	if !child.Leaf {
		child.Children = append([]*BTreeNode[T]{sibling.Children[len(sibling.Children)-1]}, child.Children...)
		sibling.Children = sibling.Children[:len(sibling.Children)-1]
	}

	n.Keys[idx-1] = sibling.Keys[len(sibling.Keys)-1]
	sibling.Keys = sibling.Keys[:len(sibling.Keys)-1]
}

func (n *BTreeNode[T]) borrowFromNext(idx int) {
	child := n.Children[idx]
	sibling := n.Children[idx+1]

	child.Keys = append(child.Keys, n.Keys[idx])
	n.Keys[idx] = sibling.Keys[0]
	sibling.Keys = sibling.Keys[1:]

	if !child.Leaf {
		child.Children = append(child.Children, sibling.Children[0])
		sibling.Children = sibling.Children[1:]
	}
}

// ---------- ПЕЧАТЬ ----------
func (n *BTreeNode[T]) Print(level int) {
	fmt.Printf("%s%v\n", spaces(level), n.Keys)
	if !n.Leaf {
		for _, child := range n.Children {
			child.Print(level + 1)
		}
	}
}

//	func spaces(level int) string {
//		return string(make([]rune, level*2))
//
// }
func spaces(level int) string {
	return strings.Repeat("  ", level) // два пробела на уровень
}
