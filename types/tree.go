package types

import "github.com/ncobase/common/validator"

// TreeNode represents a tree node.
type TreeNode interface {
	GetID() string
	GetParentID() string
	SetChildren([]TreeNode)
	GetChildren() []TreeNode
}

// BuildTree builds a tree structure from the given nodes.
func BuildTree[T TreeNode](nodes []T) []T {
	nodeMap := make(map[string]T, len(nodes))
	for _, node := range nodes {
		nodeMap[node.GetID()] = node
	}

	treeMap := make(map[string]T, len(nodes))
	for _, node := range nodeMap {
		if validator.IsEmpty(node.GetParentID()) {
			treeMap[node.GetID()] = node
		} else {
			if parent, ok := nodeMap[node.GetParentID()]; ok {
				parent.SetChildren(append(parent.GetChildren(), node))
			} else {
				treeMap[node.GetID()] = node
			}
		}
	}

	result := make([]T, 0, len(treeMap))
	for _, root := range treeMap {
		result = append(result, root)
	}

	return result
}
