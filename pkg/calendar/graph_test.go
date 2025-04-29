package calendar

import "testing"

func TestCalendarGraph(t *testing.T) {
	g1 := newGraph()

	for i := 1; i <= 6; i++ {
		g1.addVertex(int64(i))
	}

	g1.addEdge(5, 1)
	g1.addEdge(5, 2)
	g1.addEdge(5, 3)
	g1.addEdge(5, 4)

	g1.addEdge(3, 4)

	g1.addEdge(4, 2)
	g1.addEdge(4, 5)
	g1.addEdge(4, 6)

	if !g1.Nodes[5].isCyclic(g1) {
		t.Error("graph 1 has cyclic reference and was not detected")
	}

	g2 := newGraph()

	for i := 1; i <= 4; i++ {
		g2.addVertex(int64(i))
	}

	g2.addEdge(1, 2)
	g2.addEdge(1, 4)
	g2.addEdge(2, 3)
	g2.addEdge(4, 4)

	if !g2.Nodes[1].isCyclic(g2) {
		t.Error("graph 2 has cyclic reference and was not detected")
	}

	g3 := newGraph()

	for i := 1; i <= 5; i++ {
		g3.addVertex(int64(i))
	}

	g3.addEdge(1, 2)
	g3.addEdge(1, 5)
	g3.addEdge(1, 3)

	g3.addEdge(2, 5)
	g3.addEdge(5, 3)

	g3.addEdge(2, 4)

	if g3.Nodes[1].isCyclic(g3) {
		t.Error("graph 3 has no cyclic reference and was detected")
	}
}
