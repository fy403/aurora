package algorithm

import (
	"aurora/internal/tasks"
)

func TopologySort(graph *tasks.Graph) (tasks []*tasks.Signature, ok bool) {
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

func checkLoop(graph *tasks.Graph) (ok bool) {
	ok = true
	for i := 0; i < graph.VexNum; i++ {
		visited := make([]bool, graph.VexNum)
		dfs(graph, i, visited, &ok)
		if !ok {
			return
		}
	}
	return
}

func dfs(graph *tasks.Graph, v int, visited []bool, ok *bool) {
	visited[v] = true
	for w := getNeighbor(graph, v); w != -1; w = getNextNeighbor(graph, v, w) {
		if !visited[w] {
			dfs(graph, w, visited, ok)
		} else {
			*ok = false
			return
		}
	}
}

func getNeighbor(graph *tasks.Graph, v int) (w int) {
	w = -1
	for j := 0; j < graph.VexNum; j++ {
		if graph.Edge[v][j] == 1 {
			w = j
			return
		}
	}
	return
}

func getNextNeighbor(graph *tasks.Graph, v int, w int) (z int) {
	z = -1
	for j := w + 1; j < graph.VexNum; j++ {
		if graph.Edge[v][j] == 1 {
			z = j
			return
		}
	}
	return
}
