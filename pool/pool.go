package pool

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"
)

type SearchTask struct {
	FilePath string
	KeyWord  string
}
type SearchResult struct {
	FilePath string
	LinNum   int
	Content  string
}
type worker struct {
	Num int
}

type Pool struct {
	Wg       sync.WaitGroup
	workers  worker
	TasksIn  chan SearchTask
	TasksOut chan SearchResult
}

func (w *worker) Search(st *SearchTask) []SearchResult {
	file, err := os.Open(st.FilePath)
	if err != nil {
		fmt.Println("打开文件失败...")
		return nil
	}
	defer file.Close()
	var result []SearchResult
	r := bufio.NewScanner(file)
	line := 0
	for r.Scan() {
		line++
		text := r.Text()
		if strings.Contains(text, st.KeyWord) {
			result = append(result, SearchResult{
				FilePath: st.FilePath,
				LinNum:   line,
				Content:  text,
			})
		}
	}
	return result

}
func (p *Pool) Start() {
	for i := 0; i < p.workers.Num; i++ {
		p.Wg.Add(1)
		go func() {
			defer p.Wg.Done()
			for f := range p.TasksIn {
				r := p.workers.Search(&f)
				for _, s := range r {
					p.TasksOut <- s
				}
			}
		}()
	}
}
func NewPool(workerNum int, taskChanSize int) *Pool {
	return &Pool{
		workers:  worker{Num: workerNum},
		TasksIn:  make(chan SearchTask, taskChanSize),
		TasksOut: make(chan SearchResult, taskChanSize),
	}
}
