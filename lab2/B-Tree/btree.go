package B_Tree

import (
	"errors"
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

			t.Right = &TreeNode{Val: value}
			return nil
		}

		return t.Right.Insert(value)
	}

	return nil
}
