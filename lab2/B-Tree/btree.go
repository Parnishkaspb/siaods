package B_Tree

import (
	"errors"
	"golang.org/x/exp/constraints"
)

type TreeNode[T constraints.Ordered] struct {
	Val   T
	Left  *TreeNode[T]
	Right *TreeNode[T]
}

func (t *TreeNode[T]) Insert(value T) error {
	if t == nil {
		return errors.New("tree is nil")
	}

	if t.Val == value {
		return errors.New("this node value already exists")
	}

	if t.Val > value {

		if t.Left == nil {
			t.Left = &TreeNode[T]{Val: value}
			return nil
		}
		return t.Left.Insert(value)
	}

	if t.Val < value {
		if t.Right == nil {
			t.Right = &TreeNode[T]{Val: value}
			return nil
		}
		return t.Right.Insert(value)
	}

	return nil
}

func (t *TreeNode[T]) Search(value T) (*TreeNode[T], error) {
	if t == nil {
		return nil, errors.New("tree is nil")
	}

	switch {
	case value == t.Val:
		return t, nil
	case value < t.Val:
		return t.Left.Search(value)
	default:
		return t.Right.Search(value)
	}
}

func (t *TreeNode[T]) Delete(value T) {
	t.remove(value)
}

func (t *TreeNode[T]) remove(value T) (*TreeNode[T], error) {
	if t == nil {
		return nil, errors.New("tree is nil")
	}

	if value < t.Val {
		t.Left, _ = t.Left.remove(value)
		return t, nil
	}
	if value > t.Val {
		t.Right, _ = t.Right.remove(value)
		return t, nil
	}

	if t.Left == nil && t.Right == nil {
		t = nil
		return nil, errors.New("tree is nil")
	}

	if t.Left == nil {
		t = t.Right
		return t, nil
	}
	if t.Right == nil {
		t = t.Left
		return t, nil
	}

	smallestValOnRight := t.Right
	for {
		if smallestValOnRight != nil && smallestValOnRight.Left != nil {
			smallestValOnRight = smallestValOnRight.Left
		} else {
			break
		}
	}

	t.Val = smallestValOnRight.Val
	t.Right, _ = t.Right.remove(t.Val)
	return t, nil
}

func (t *TreeNode[T]) FindMax() any {
	if t.Right == nil {
		return t.Val
	}
	return t.Right.FindMax()
}

func (t *TreeNode[T]) FindMin() any {
	if t.Left == nil {
		return t.Val
	}
	return t.Left.FindMin()
}
