package types

import (
	"ncobase/common/validator"
	"sort"
)

type SortField string

// TreeNode represents a tree node.
type TreeNode interface {
	GetID() string
	GetParentID() string
	SetChildren([]TreeNode)
	GetChildren() []TreeNode
	GetSortValue(field string) any
}

// BuildTree builds a tree structure from the given nodes.
func BuildTree[T TreeNode](nodes []T, sortField string) []T {
	// Sort all nodes before building the tree
	sort.SliceStable(nodes, func(i, j int) bool {
		return compareSortValues(nodes[i].GetSortValue(sortField), nodes[j].GetSortValue(sortField))
	})

	nodeMap := make(map[string]T, len(nodes))
	for _, node := range nodes {
		nodeMap[node.GetID()] = node
	}

	var roots []T
	for _, node := range nodes {
		if validator.IsEmpty(node.GetParentID()) {
			roots = append(roots, node)
		} else {
			if parent, ok := nodeMap[node.GetParentID()]; ok {
				parent.SetChildren(append(parent.GetChildren(), node))
			} else {
				roots = append(roots, node)
			}
		}
	}

	// Sort children for each node
	for _, node := range nodeMap {
		sortChildrenGeneric(node, sortField)
	}

	return roots
}

// sortChildrenGeneric sort children node
func sortChildrenGeneric(node TreeNode, sortField string) {
	children := node.GetChildren()
	sort.SliceStable(children, func(i, j int) bool {
		return compareSortValues(children[i].GetSortValue(sortField), children[j].GetSortValue(sortField))
	})
	node.SetChildren(children)

	// Recursively sort children of children
	for _, child := range children {
		sortChildrenGeneric(child, sortField)
	}
}

// compareSortValues compare sort values
func compareSortValues(a, b any) bool {
	switch va := a.(type) {
	case int:
		return va < b.(int)
	case int64:
		return va < b.(int64)
	case float64:
		return va < b.(float64)
	case string:
		return va < b.(string)
	default:
		return false
	}
}
