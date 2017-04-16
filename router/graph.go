package router

var closed = make(chan int)

func init() {
	close(closed)
}

// Graph represents a directed graph
type Graph interface {
	// Connect creates directed edge between vertices head and tail.
	Connect(head, tail int)

	// Disconnect removes directed edge between vertices head and tail.
	Disconnect(head, tail int)

	// Neighbors returns channel of all direct predecessors of vertex head.
	Neighbors(head int) <-chan int
}

// "adjacency list" with O(1) amortized connect/disconnect time
type sparseGraph map[int]map[int]struct{}

func (g sparseGraph) Connect(head, tail int) {
	vertices, ok := g[head]
	if !ok {
		vertices = make(map[int]struct{}, 1)
		g[head] = vertices
	}
	vertices[tail] = struct{}{}
}

func (g sparseGraph) Disconnect(head, tail int) {
	if vertices, ok := g[head]; ok {
		delete(vertices, tail)
		if len(vertices) == 0 {
			delete(g, head)
		}
	}
}

func (g sparseGraph) Neighbors(head int) <-chan int {
	if pred, ok := g[head]; ok {
		c := make(chan int)
		go func() {
			for v := range pred {
				c <- v
			}
			close(c)
		}()
		return c
	}
	return closed
}

// adjacency matrix with true O(1) connect/disconnect time
type denseGraph struct {
	bytes []byte
	maxN  int
}

func (g denseGraph) resize(n int) {
	// TODO: implement
}

func (g denseGraph) Connect(head, tail int) {
	// TODO: implement
}

func (g denseGraph) Disconnect(head, tail int) {
	// TODO: implement
}
