package vi

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"net/http"
)

// helper to check the  where two paths diverge , so at that point more node will be create
// therefore node to be create =  len(newP) - level
func divergeAt(oldP, newP string) (level int) {
	cmap := make(map[rune]int)

	for _, rs := range oldP + newP {
		cmap[rs]++
	}

	for i, v := range newP {
		if cmap[v] == 1 {
			// Found diverge
			return i
		}
	}
	return -1
}

var _ = Describe("tree: unix-test", Ordered, func() {
	var (
		nTree               *tree
		root                *treenode
		nTreeSizeInitial    int
		nTreeSizeAfterFirst int
		handler             http.HandlerFunc
		path1               = "/hello"
		path2               = "/hello2"
	)
	BeforeAll(func() {
		handler = func(w http.ResponseWriter, r *http.Request) {

			fmt.Fprint(w, "ok")
		}
		nTree = newTree()
		root = nTree.root
		Expect(nTree).To(Not(BeNil()), "tree should not be nil when initialize")
		Expect(nTree.size).To(Equal(1), "initial tree size should be 1")
		Expect(nTree.root.depth).To(Equal(1), "initial root depth should be 1")
		Expect(nTree.root.children).To(BeEmpty(), "initial children of root should be 0")
		nTreeSizeInitial = nTree.size

	})

	When("Adding node , it should be reuse , added value  and order correctly to the tree", func() {
		When("Adding initial path", func() {
			path1Len := len(path1)
			mock := setupMock(handler, path1)

			It(fmt.Sprintf("Should create %d nodes when adding path : %s", path1Len, path1), func() {

				nTree.add(path1, mock.GetHandler())
				currentNtreeSize := nTree.size
				diffLevel := divergeAt(path1, string(root.key))
				if diffLevel < 0 {
					// no diverge , so len(path) of nodes need to be create
					diffLevel = len(path1)
				}
				expectSize := nTreeSizeInitial + len(path1) - diffLevel

				Expect(currentNtreeSize).To(Equal(expectSize), "Tree size should be %d after create %d nodes for path : %s", expectSize, len(path1)-diffLevel, path1)
				nTreeSizeAfterFirst = currentNtreeSize
			})
			It("Added correctly", func() {
				node := root

				for i, key := range path1 {
					if !node.isLeaf {
						By("Create correct amount of  node")
						Expect(node.children).To(HaveLen(1), fmt.Sprintf("Node %v should have a single children with %v \n", string(node.key), key))

						By("Only the final node will be leaf node ")
						Expect(node.isLeaf).To(BeFalse(), "node %v is not leaf node", string(node.key))
						Expect(node.path).To(BeEmpty(), "node %v is not leaf node therefore path should be empty", string(node.key))

						By("Added correct childrens")
						Expect(node.depth).To(Equal(i+1), "the depth level of node %v should be %d", string(node.key), i+1)
						children, childrenExist := node.children[nodeKey(key)]
						Expect(childrenExist).To(BeTrue(), "the children of node %v should exist", string(node.key))

						node = children
					}

				}

				hlr := node.handler
				Expect(hlr).ToNot(BeNil(), "leaf node should contain added handler")

			})
		})

		When("Adding second path", func() {
			mock := setupMock(handler, path2)

			It("Should create the different (between len(path)  and the different of two paths) nodes only", func() {
				nTree.add(path2, mock.GetHandler())
				currentSize := nTree.size
				diffLevel := divergeAt(path1, path2)
				if diffLevel < 0 {
					diffLevel = len(path2)
				}

				expectSize := nTreeSizeAfterFirst + len(path2) - diffLevel

				Expect(currentSize).To(Equal(expectSize), "Tree size should be %d after create %d nodes for path : %s", expectSize, len(path2)-diffLevel, path2)
			})

			It("Should added correctly", func() {
				node := root

				diffLevel := divergeAt(path1, path2)

				for i, key := range path2 {
					if !node.isLeaf {
						By("Reuse node")
						if i == diffLevel {
							Expect(node.children).To(HaveLen(2), fmt.Sprintf("Node %d of key %s should be reuse, and therfore another children should be added", diffLevel, string(path2[diffLevel-1])))
						} else {

							Expect(node.children).To(HaveLen(1), "other newly added node will  have children of 1")
						}

						By("Only the final node will be leaf node ")
						Expect(node.isLeaf).To(BeFalse(), "node %v is not leaf node", string(node.key))
						Expect(node.path).To(BeEmpty(), "node %v is not leaf node therefore path should be empty", string(node.key))

						By("Added correct childrens")
						Expect(node.depth).To(Equal(i+1), "the depth level of node %v should be %d", string(node.key), i+1)
						children, childrenExist := node.children[nodeKey(key)]
						Expect(childrenExist).To(BeTrue(), "the children of node %v should exist", string(node.key))

						node = children

					}

				}

				hlr := node.handler
				Expect(hlr).ToNot(BeNil(), "leaf node should contain added handler")
			})
		})

	})

	When("Finding node in the tree", func() {
		It("Should return matches node", func() {
			By("Finding nodes for the first path")
			nodes := nTree.find(path1)
			Expect(nodes).ToNot(BeNil(), fmt.Sprintf("nodes for %s should not be nil", path1))
			Expect(nodes).To(HaveLen(2), "expect path hello to match both handler for hello and hello2")
			Expect(string(nodes[0].key)).To(Equal("o"), "last element of path1 is o")
			By("Finding nodes for the second path")
			nodes = nTree.find(path2)
			Expect(nodes).ToNot(BeNil(), fmt.Sprintf("nodes for %s should not be nil", path2))
			Expect(nodes).To(HaveLen(1), "expect path hello2 to match only handler for hello2 not hello")
			Expect(string(nodes[0].key)).To(Equal("2"))
			By("Finding nodes for invalid path")
			nodes = nTree.find("invalid")
			Expect(nodes).To(BeEmpty(), "should not return any nodes for invalid path")
		})

	})

})
