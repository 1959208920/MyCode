package scheduler

import (
	"math"
	"time"
)

func ReorderSort(graph, invgraph [][]int32) ([]int32, error) {
	indegree := make(map[int32]int)
	outdegree := make(map[int32]int)
	nodeset := make(map[int32]bool)

	start1 := time.Now()
	// in-degree
	for i := 0; i < len(invgraph); i++ {
		for node := range invgraph {
			indegree[int32(node)] = len(invgraph[int32(node)])
			nodeset[int32(node)] = true
		}
	}
	// out-degree
	for i := 0; i < len(graph); i++ {
		for node := range graph {
			outdegree[int32(node)] = len(graph[int32(node)])
		}
	}

	var result []int32
	var nextSort int32

	start2 := time.Now()
	for len(nodeset) > 0 {
		// find nextDeleteNode
		minIndegree := math.MaxInt32
		for node := range nodeset {
			if indegree[node] < minIndegree {
				minIndegree = indegree[node]
				nextSort = node
			} else if indegree[node] == minIndegree && outdegree[node] < outdegree[nextSort] {
				nextSort = node
			}
		}

		// remove all related edges
		for _, nextDelete := range invgraph[nextSort] {
			if exist := nodeset[nextDelete]; !exist {
				continue
			}
			delete(nodeset, nextDelete)
			for _, v := range invgraph[nextDelete] {
				outdegree[v]--
			}
			for _, v := range graph[nextDelete] {
				indegree[v]--
			}
		}
		result = append(result, nextSort)
		for _, v := range graph[nextSort] {
			indegree[v]--
		}
		delete(nodeset, nextSort)
	}

	return result, nil
}
