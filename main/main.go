package main

import (
	"fmt"
	"os"
	"path/filepath"
	"searching/pool"
)

var dir string
var keyword string

const (
	poolnum  = 5
	poolchan = 100
)

var workers *pool.Pool

func makew() {
	workers = pool.NewPool(poolnum, poolchan)
}

func startw() {
	workers.Start()
}

// 这段递归靠的ai，我尽量看懂先
func WorkPush(filepatht string, keywordt string) {
	fmt.Println("开始递归搜索目录:", filepatht)
	//这里应该用了递归，但我实在没看出来，可能是func这里，"path/filepath"主要是不知道这个包怎么用
	err := filepath.Walk(filepatht, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		//这个大概是找到了文件而不是目录，所以传入管道
		if !info.IsDir() {
			task := pool.SearchTask{
				FilePath: path,
				KeyWord:  keywordt,
			}
			// 发送任务给 Worker
			workers.TasksIn <- task
		}
		//没找到返回nil
		return nil
	})

	if err != nil {
		fmt.Printf("遍历目录出错: %v\n", err)
	}
}

// 这里问的，不过能理解
// 3. 原理解析
// os.Args: 这是一个 []string 切片。
// 当你在终端输入 ./catch ./ func 并回车时：
// os.Args[0] = "./catch"
// os.Args[1] = "./"
// os.Args[2] = "func"
// 这就是为什么我们要判断 len(os.Args) < 3，防止用户只输了命令没输参数导致数组越界 panic。
func inputThing() (string, string) {
	if len(os.Args) < 3 { //参数少了返回true，maybe
		fmt.Println("使用方法错误！")
		fmt.Println("正确格式: ./catch [目录] [关键词]")
		fmt.Println("示例: ./catch ./src func")
		os.Exit(1)
	}
	//这里应该是通过os的Args方法返回输入的值
	d := os.Args[1]
	k := os.Args[2]
	fmt.Printf("正在搜索目录: %s , 关键词: %s \n", d, k)
	return d, k
}
func main() {
	dir, keyword = inputThing()
	makew()
	startw()
	done := make(chan struct{})
	go func() {
		for to := range workers.TasksOut {
			fmt.Printf("路径: %s , 行数: %d , 内容: %s\n", to.FilePath, to.LinNum, to.Content)
		}
		close(done)
	}()
	WorkPush(dir, keyword)
	close(workers.TasksIn)
	workers.Wg.Wait()
	close(workers.TasksOut)
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
