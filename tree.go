package vi

import (
	"net/http"
	"strings"
)

type (
	nodeKey = string

	// represent the tree of nodes for particular route method
	tree struct {
		// tree root
		root *treenode
		// tree size
		size int
	}

	// tree node represent the structure that hold particular identify by particular node key
	// only leaf node  will hold the expected value in this case handler and middlewares
	treenode struct {
		key nodeKey
		// the route
		path string
		// the handler for the router
		handler http.HandlerFunc
		// its depth in the routing tree
		depth int
		// its children treenodes
		children map[nodeKey]*treenode
		// whether treenode is leaf treenode
		isLeaf bool
		// middlewares register on the node
		prefixes []string
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
		root: newNode("/", 1),
		size: 1,
	}
}

// add new route to the routing tree
// create a new treenode for each char in the key
// and the final treenode represent the endpoint of the route
func (tree *tree) add(path string, handle http.HandlerFunc, prefixes []string) {
	var treenode = tree.root

	if path != treenode.key {
		path = strings.TrimPrefix(path, "/")

		for _, key := range path {
			child, ok := treenode.children[nodeKey(key)]
			if !ok {
				child = newNode(nodeKey(key), treenode.depth+1)
				treenode.children[nodeKey(key)] = child
				tree.size++
			}

			treenode = child
		}

		path = "/" + path
	}

	treenode.handler = handle
	treenode.isLeaf = true
	treenode.path = path
	treenode.prefixes = prefixes
}

// find all nodes in the routing tree that match given key
// After identify  the first matching , it continue to find  all leaf nodes with  bread first exploration
func (tree *tree) find(key string) (nodes []*treenode) {
	var (
		stack []*treenode
		node  = tree.root
	)

	// root
	if key == node.path {
		return []*treenode{node}
	}

	keys := strings.TrimPrefix(key, "/")

	for _, char := range keys {
		child, ok := node.children[nodeKey(char)]
		if !ok {
			return
		}
		if key == child.path {
			return []*treenode{child}
		}

		node = child
	}

	// successfully match the entire key
	stack = append(stack, node)

	for len(stack) > 0 {
		node, stack = stack[0], stack[1:]

		if node.isLeaf {
			nodes = append(nodes, node)
		}

		for _, vnode := range node.children {
			stack = append(stack, vnode)
		}
	}

	return nodes
}
