package R_Tree

import (
	"math"
	"sort"
)

const maxEntries = 4
const minEntries = 2

type Rect struct {
	MinX, MinY, MaxX, MaxY float64
}

func (r Rect) Intersects(o Rect) bool {
	return r.MinX <= o.MaxX && r.MaxX >= o.MinX &&
		r.MinY <= o.MaxY && r.MaxY >= o.MinY
}

func (r Rect) Area() float64 {
	return (r.MaxX - r.MinX) * (r.MaxY - r.MinY)
}

func (r Rect) Union(o Rect) Rect {
	return Rect{
		MinX: math.Min(r.MinX, o.MinX),
		MinY: math.Min(r.MinY, o.MinY),
		MaxX: math.Max(r.MaxX, o.MaxX),
		MaxY: math.Max(r.MaxY, o.MaxY),
	}
}

func (r Rect) Contains(o Rect) bool {
	return r.MinX <= o.MinX && r.MaxX >= o.MaxX &&
		r.MinY <= o.MinY && r.MaxY >= o.MaxY
}

func (r Rect) DistanceToPoint(x, y float64) float64 {
	dx := math.Max(0, math.Max(r.MinX-x, x-r.MaxX))
	dy := math.Max(0, math.Max(r.MinY-y, y-r.MaxY))
	return math.Hypot(dx, dy)
}

type Item struct {
	Rect Rect
	Data any
}

type Node struct {
	Leaf     bool
	Children []*Node
	Items    []Item
	Bounds   Rect
}

type RTree struct {
	Root *Node
}

func New() *RTree {
	return &RTree{
		Root: &Node{Leaf: true},
	}
}

func (t *RTree) Insert(item Item) {
	n := t.chooseLeaf(t.Root, item.Rect)
	n.Items = append(n.Items, item)
	n.Bounds = expandRect(n.Bounds, item.Rect)

	if len(n.Items) > maxEntries {
		t.splitNode(n)
	}
}

func (t *RTree) Delete(targetData any) bool {
	found, newRoot := deleteRecursive(t.Root, targetData)
	if found {
		t.Root = newRoot
	}
	return found
}

func (t *RTree) Search(query Rect) []Item {
	return search(t.Root, query)
}

func (t *RTree) Knn(x, y float64, k int) []Item {
	var result []Item
	collectKnn(t.Root, x, y, k, &result)
	return result
}

func (t *RTree) chooseLeaf(n *Node, r Rect) *Node {
	if n.Leaf {
		return n
	}
	var best *Node
	bestEnlargement := math.MaxFloat64
	for _, child := range n.Children {
		union := child.Bounds.Union(r)
		enlargement := union.Area() - child.Bounds.Area()
		if enlargement < bestEnlargement {
			best = child
			bestEnlargement = enlargement
		}
	}
	return t.chooseLeaf(best, r)
}

func (t *RTree) splitNode(n *Node) {
	if n.Leaf {
		t.splitLeaf(n)
	} else {
		t.splitBranch(n)
	}
}

func (t *RTree) splitLeaf(n *Node) {
	total := len(n.Items)
	mid := total / 2
	if mid < minEntries {
		mid = minEntries
	}
	if total-mid < minEntries {
		mid = total - minEntries
	}
	left := &Node{Leaf: true, Items: n.Items[:mid]}
	right := &Node{Leaf: true, Items: n.Items[mid:]}
	left.Bounds = computeItemsBounds(left.Items)
	right.Bounds = computeItemsBounds(right.Items)
	*n = Node{Leaf: false, Children: []*Node{left, right}, Bounds: left.Bounds.Union(right.Bounds)}
}

func (t *RTree) splitBranch(n *Node) {
	total := len(n.Children)
	mid := total / 2
	if mid < minEntries {
		mid = minEntries
	}
	if total-mid < minEntries {
		mid = total - minEntries
	}
	left := &Node{Leaf: false, Children: n.Children[:mid]}
	right := &Node{Leaf: false, Children: n.Children[mid:]}
	left.Bounds = computeChildrenBounds(left.Children)
	right.Bounds = computeChildrenBounds(right.Children)
	*n = Node{Leaf: false, Children: []*Node{left, right}, Bounds: left.Bounds.Union(right.Bounds)}
}

func deleteRecursive(n *Node, target any) (bool, *Node) {
	if n.Leaf {
		var newItems []Item
		found := false
		for _, item := range n.Items {
			if item.Data == target {
				found = true
				continue
			}
			newItems = append(newItems, item)
		}
		if !found {
			return false, n
		}
		if len(newItems) < minEntries {
			return true, nil
		}
		n.Items = newItems
		n.Bounds = computeItemsBounds(newItems)
		return true, n
	}
	var newChildren []*Node
	changed := false
	for _, child := range n.Children {
		found, newChild := deleteRecursive(child, target)
		if found {
			changed = true
		}
		if newChild != nil {
			newChildren = append(newChildren, newChild)
		}
	}
	if !changed {
		return false, n
	}
	if len(newChildren) < minEntries {
		return true, nil
	}
	if len(newChildren) == 1 {
		return true, newChildren[0]
	}
	n.Children = newChildren
	n.Bounds = computeChildrenBounds(newChildren)
	return true, n
}

func search(n *Node, r Rect) []Item {
	var result []Item
	if !n.Bounds.Intersects(r) {
		return result
	}
	if n.Leaf {
		for _, item := range n.Items {
			if item.Rect.Intersects(r) {
				result = append(result, item)
			}
		}
	} else {
		for _, child := range n.Children {
			result = append(result, search(child, r)...)
		}
	}
	return result
}

func collectKnn(n *Node, x, y float64, k int, result *[]Item) {
	if n == nil {
		return
	}
	if n.Leaf {
		for _, item := range n.Items {
			dist := item.Rect.DistanceToPoint(x, y)
			*result = append(*result, Item{Rect: item.Rect, Data: struct {
				any
				Dist float64
			}{item.Data, dist}})
		}
		sort.Slice(*result, func(i, j int) bool {
			return (*result)[i].Data.(struct {
				any
				Dist float64
			}).Dist < (*result)[j].Data.(struct {
				any
				Dist float64
			}).Dist
		})
		if len(*result) > k {
			*result = (*result)[:k]
		}
	} else {
		var childrenWithDist []struct {
			Child *Node
			Dist  float64
		}
		for _, child := range n.Children {
			childrenWithDist = append(childrenWithDist, struct {
				Child *Node
				Dist  float64
			}{child, child.Bounds.DistanceToPoint(x, y)})
		}
		sort.Slice(childrenWithDist, func(i, j int) bool {
			return childrenWithDist[i].Dist < childrenWithDist[j].Dist
		})
		for _, entry := range childrenWithDist {
			collectKnn(entry.Child, x, y, k, result)
		}
	}
}

func computeItemsBounds(items []Item) Rect {
	if len(items) == 0 {
		return Rect{}
	}
	b := items[0].Rect
	for _, item := range items[1:] {
		b = b.Union(item.Rect)
	}
	return b
}

func computeChildrenBounds(children []*Node) Rect {
	if len(children) == 0 {
		return Rect{}
	}
	b := children[0].Bounds
	for _, c := range children[1:] {
		b = b.Union(c.Bounds)
	}
	return b
}

func expandRect(a, b Rect) Rect {
	if a.Area() == 0 {
		return b
	}
	return a.Union(b)
}
