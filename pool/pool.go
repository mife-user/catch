package pool

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type SearchTask struct {
	FilePath   string
	KeyWord    string
	SearchType bool
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
	var result []SearchResult
	if st.SearchType {
		if strings.Contains(st.FilePath, st.KeyWord) {
			result = append(result, SearchResult{
				FilePath: st.FilePath,
				LinNum:   0,
				Content:  "Matches Filename",
			})
		}
		return result
	}
	file, err := os.Open(st.FilePath)
	if err != nil {
		fmt.Println("打开文件失败...")
		return nil
	}
	defer file.Close()

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
func (p *Pool) WorkPush(filepatht string, keywordt string, isFileSearching bool) {
	fmt.Println("开始递归搜索目录:", filepatht)
	//这里应该用了递归，但我实在没看出来，可能是func这里，"path/filepath"主要是不知道这个包怎么用
	err := filepath.WalkDir(filepatht, func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && !isFileSearching {
			return nil
		} else {
			task := SearchTask{
				FilePath:   path,
				KeyWord:    keywordt,
				SearchType: isFileSearching,
			}
			// 发送任务给 Worker
			p.TasksIn <- task
		}

		//没找到返回nil
		return nil
	})

	if err != nil {
		fmt.Printf("遍历目录出错: %v\n", err)
	}
}
