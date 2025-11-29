package main

import (
	"flag"
	"fmt"
	"os"
	"searching/pool"
)

const (
	poolnum  = 5
	poolchan = 100
)

func inputThing() (string, string, bool) {
	var searchfile bool
	flag.BoolVar(&searchfile, "v", false, "查找目录名")
	flag.Usage = func() {
		fmt.Println("使用方式: ./catch [-v] [目录] [关键词]")
		fmt.Println("说明: -v 选项必须放在目录前面")
		flag.PrintDefaults()
	}
	flag.Parse()
	args := flag.Args()
	if len(args) < 2 { //参数少了返回true，maybe
		fmt.Println("使用方法错误！")
		fmt.Println("正确格式: ./catch [目录] [关键词] [模式]")
		fmt.Println("示例: ./catch ./src func -v")
		os.Exit(1)
	}
	d := args[0]
	k := args[1]
	if searchfile {
		fmt.Printf("正在搜索目录：%s,关键词：%s\n", d, k)
		return d, k, searchfile
	}
	fmt.Printf("正在搜索内容: %s , 关键词: %s \n", d, k)
	return d, k, searchfile
}
func main() {
	dir, keyword, function := inputThing()
	p := pool.NewPool(poolnum, poolchan)
	p.Start()
	done := make(chan struct{})
	go func() {
		for to := range p.TasksOut {
			fmt.Printf("路径: %s\n , 行数: %d\n , 内容: %s\n", to.FilePath, to.LinNum, to.Content)
		}
		close(done)
	}()
	p.WorkPush(dir, keyword, function)
	close(p.TasksIn)
	p.Wg.Wait()
	close(p.TasksOut)
	<-done
	fmt.Println("game over>.<")
}

//当笔记用：
/*2. 知识点讲解：filepath.Walk 到底是怎么递归的？
你疑惑：“这里应该用了递归，但我实在没看出来”。
其实，递归逻辑被封装在 filepath.Walk 这个函数内部了。你写的那个 func 叫做回调函数 (Callback)。
通俗类比：
filepath.Walk 是一个扫地机器人。
你写的那个 func 是你贴在机器人脑门上的便签纸，上面写着：“每进一个房间，或者看到一张纸，就执行这些指令”。
流程如下：
你调用 Walk("A目录", 你的func)。
机器人进入 A 目录。
机器人捡起第一个文件 a.txt -> 调用你的func (告诉你：我找到 a.txt 了，不是目录)。
你的 func 判断：不是目录 -> 塞入管道。
机器人发现有个子目录 B文件夹 -> 调用你的func (告诉你：我找到 B 了，是目录)。
你的 func 判断：是目录 -> 不做处理。
机器人自动进入 B文件夹 (这就是 Walk 内部的递归)。
重复上述过程，直到走完所有角落。*/
