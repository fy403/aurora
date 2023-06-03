package tasks

func TopologySort(graph *Graph) (tasks []*Signature, ok bool) {
	// Check for loop
	if !checkLoop(graph) {
		return nil, false
	}

	// Generate initial executed sequences
	inDegrees := make([]int, graph.VexNum)
	for i := 0; i < graph.VexNum; i++ {
		for j := 0; j < graph.VexNum; j++ {
			if graph.Edge[i][j] == 1 {
				inDegrees[j]++
			}
		}
	}

	// Initial sequences
	for v, n := range inDegrees {
		if n == 0 {
			tasks = append(tasks, graph.Vertexes[v])
		}
	}

	return tasks, true
}

func checkLoop(graph *Graph) (ok bool) {
	ok = true
	visited := make([]bool, graph.VexNum)
	for i := 0; i < graph.VexNum; i++ {
		for idx := range visited {
			visited[idx] = false
		}
		dfs(graph, i, visited, &ok)
	}
	return
}

func dfs(graph *Graph, v int, visited []bool, ok *bool) {
	visited[v] = true
	for w := getNeighbor(graph, v); w != -1; w = getNextNeighbor(graph, v, w) {
		if !visited[w] {
			dfs(graph, w, visited, ok)
		} else {
			*ok = false
			return
		}
	}
	visited[v] = false
}

func getNeighbor(graph *Graph, v int) (w int) {
	w = -1
	for j := 0; j < graph.VexNum; j++ {
		if graph.Edge[v][j] == 1 {
			w = j
			return
		}
	}
	return
}

func getNextNeighbor(graph *Graph, v int, w int) (z int) {
	z = -1
	for j := w + 1; j < graph.VexNum; j++ {
		if graph.Edge[v][j] == 1 {
			z = j
			return
		}
	}
	return
}
