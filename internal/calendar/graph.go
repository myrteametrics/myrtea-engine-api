package calendar

//Node ..
type Node struct {
	ID    int64
	Nodes []*Node
}

//Graph ..
type Graph struct {
	Nodes    map[int64]*Node
	Visited  map[int64]bool
	RecStack map[int64]bool
}

func (n *Node) contains(other *Node) bool {
	for _, sub := range n.Nodes {
		if sub.ID == other.ID {
			return true
		}
	}
	return false
}

func (n *Node) isCyclic(graph *Graph) bool {

	if graph.RecStack[n.ID] {
		return true
	}
	if graph.Visited[n.ID] {
		return false
	}

	graph.Visited[n.ID] = true
	graph.RecStack[n.ID] = true

	for _, nn := range n.Nodes {
		if nn.isCyclic(graph) {
			return true
		}
	}

	graph.RecStack[n.ID] = false

	return false
}

func (graph *Graph) addVertex(id int64) bool {
	_, found := graph.Nodes[id]

	if !found {
		graph.Nodes[id] = &Node{ID: id, Nodes: make([]*Node, 0)}
		graph.RecStack[id] = false
		graph.Visited[id] = false
		return true
	}

	return false
}

func (graph *Graph) clearGraph() {
	for id := range graph.RecStack {
		graph.RecStack[id] = false
	}

	for id := range graph.Visited {
		graph.Visited[id] = false
	}
}

func (graph *Graph) addEdge(from int64, to int64) bool {
	nodeFrom, ok1 := graph.Nodes[from]
	nodeTo, ok2 := graph.Nodes[to]

	if !ok1 || !ok2 || nodeFrom.contains(nodeTo) {
		return false
	}

	nodeFrom.Nodes = append(nodeFrom.Nodes, nodeTo)

	return true
}

func newGraph() *Graph {
	return &Graph{
		Nodes:    map[int64]*Node{},
		RecStack: map[int64]bool{},
		Visited:  map[int64]bool{},
	}
}
