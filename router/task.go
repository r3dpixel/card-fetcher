package router

import "github.com/r3dpixel/card-fetcher/task"

type TaskBucket struct {
	Tasks       map[string]task.Task
	ValidURLs   []string
	InvalidURLs []string
}

type TaskSlice struct {
	Tasks       []task.Task
	ValidURLs   []string
	InvalidURLs []string
}
