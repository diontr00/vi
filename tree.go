package vi

import (
	"net/http"
)

type (
	nodeKey = rune

	tree struct {
		// tree root
		root *treenode
		// tree size
		size int
	}

	treenode struct {
		key nodeKey
		// the route
		path string
		// the handler for the router
		handler http.HandlerFunc
		// its depth in the routing tree
		depth int
		// its children treenode
		children map[nodeKey]*treenode
		// whether treenode is leaf treenode
		isLeaf bool
	}
)

// create a new treenode with specific key
func newNode(key nodeKey, depth int) *treenode {
	return &treenode{
		key:      key,
		depth:    depth,
		children: make(map[nodeKey]*treenode),
	}
}

func newTree() *tree {
	return &tree{
		root: newNode(' ', 1),
		size: 1,
	}
}

// add new route to the routing tree
// create a new treenode for each char in the key
// and the final treenode represent the endpoint of the route
func (tree *tree) add(path string, handle http.HandlerFunc) {
	var treenode = tree.root

	for _, char := range path {
		child, ok := treenode.children[nodeKey(char)]
		if !ok {
			child = newNode(nodeKey(char), treenode.depth+1)
			treenode.children[nodeKey(char)] = child
			tree.size++
		}

		treenode = child
	}

	treenode.handler = handle
	treenode.isLeaf = true
	treenode.path = path
}

// find all nodes in the routing tree that match given key
// After identify  the first matching , it continue to find  all leaf nodes with  bread first exploration
func (tree *tree) find(key string) (nodes []*treenode) {
	var (
		queue []*treenode
		node  = tree.root
	)

	for _, char := range key {
		child, ok := node.children[nodeKey(char)]
		if !ok {
			return
		}
		node = child
	}

	// successfully match the entire key
	queue = append(queue, node)
	for len(queue) > 0 {
		var tmpQueue []*treenode
		for _, node := range queue {
			if node.isLeaf {
				nodes = append(nodes, node)
			}

			for _, vnode := range node.children {
				tmpQueue = append(tmpQueue, vnode)
			}
		}
		queue = tmpQueue
	}
	return
}
