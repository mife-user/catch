package search

import (
	"encoding/csv"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// TestPageResults 测试分页功能
func TestPageResults(t *testing.T) {
	// 创建测试结果
	results := make([]SearchResult, 25)
	for i := 0; i < 25; i++ {
		results[i] = SearchResult{
			FilePath:  filepath.Join("test", "file.go"),
			Lines:     []string{"test line"},
			MatchType: "content",
		}
	}

	tests := []struct {
		name         string
		results      []SearchResult
		page         int
		pageSize     int
		wantPage     int
		wantTotal    int
		wantLen      int
		wantHasNext  bool
		wantHasPrev  bool
	}{
		{
			name:        "正常分页 - 第1页",
			results:     results,
			page:        1,
			pageSize:    10,
			wantPage:    1,
			wantTotal:   3,
			wantLen:     10,
			wantHasNext: true,
			wantHasPrev: false,
		},
		{
			name:        "正常分页 - 第2页",
			results:     results,
			page:        2,
			pageSize:    10,
			wantPage:    2,
			wantTotal:   3,
			wantLen:     10,
			wantHasNext: true,
			wantHasPrev: true,
		},
		{
			name:        "正常分页 - 最后一页",
			results:     results,
			page:        3,
			pageSize:    10,
			wantPage:    3,
			wantTotal:   3,
			wantLen:     5,
			wantHasNext: false,
			wantHasPrev: true,
		},
		{
			name:        "空结果",
			results:     []SearchResult{},
			page:        1,
			pageSize:    10,
			wantPage:    1,
			wantTotal:   0,
			wantLen:     0,
			wantHasNext: false,
			wantHasPrev: false,
		},
		{
			name:        "单条结果",
			results:     results[:1],
			page:        1,
			pageSize:    10,
			wantPage:    1,
			wantTotal:   1,
			wantLen:     1,
			wantHasNext: false,
			wantHasPrev: false,
		},
		{
			name:        "页码超出范围",
			results:     results,
			page:        10,
			pageSize:    10,
			wantPage:    3,
			wantTotal:   3,
			wantLen:     5,
			wantHasNext: false,
			wantHasPrev: true,
		},
		{
			name:        "pageSize 为 0",
			results:     results,
			page:        1,
			pageSize:    0,
			wantPage:    1,
			wantTotal:   3,
			wantLen:     10, // 默认 pageSize 为 10
			wantHasNext: true,
			wantHasPrev: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			paged := PageResults(tt.results, tt.page, tt.pageSize)

			if paged.Page != tt.wantPage {
				t.Errorf("Page = %d, want %d", paged.Page, tt.wantPage)
			}
			if paged.TotalPages != tt.wantTotal {
				t.Errorf("TotalPages = %d, want %d", paged.TotalPages, tt.wantTotal)
			}
			if paged.TotalCount != len(tt.results) {
				t.Errorf("TotalCount = %d, want %d", paged.TotalCount, len(tt.results))
			}
			if len(paged.Results) != tt.wantLen {
				t.Errorf("Results length = %d, want %d", len(paged.Results), tt.wantLen)
			}
			if paged.HasNext != tt.wantHasNext {
				t.Errorf("HasNext = %v, want %v", paged.HasNext, tt.wantHasNext)
			}
			if paged.HasPrev != tt.wantHasPrev {
				t.Errorf("HasPrev = %v, want %v", paged.HasPrev, tt.wantHasPrev)
			}
		})
	}
}

// TestContainsIgnoreCase 测试大小写无关匹配
func TestContainsIgnoreCase(t *testing.T) {
	tests := []struct {
		name  string
		s     string
		sub   string
		want  bool
	}{
		{"基本匹配", "hello world", "world", true},
		{"大小写混合", "Hello World", "WORLD", true},
		{"全部大写", "HELLO WORLD", "hello", true},
		{"全部小写", "hello world", "HELLO", true},
		{"不匹配", "hello world", "foo", false},
		{"空子串", "hello world", "", true},
		{"空字符串", "", "hello", false},
		{"都空", "", "", true},
		{"部分匹配", "abcdef", "bcd", true},
		{"单个字符", "abc", "b", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := containsIgnoreCase(tt.s, tt.sub)
			if got != tt.want {
				t.Errorf("containsIgnoreCase(%q, %q) = %v, want %v", tt.s, tt.sub, got, tt.want)
			}
		})
	}
}

// TestShouldSkipDir 测试目录跳过
func TestShouldSkipDir(t *testing.T) {
	tests := []struct {
		name string
		dir  string
		want bool
	}{
		{"git 目录", ".git", true},
		{"node_modules", "node_modules", true},
		{"vendor", "vendor", true},
		{"vscode", ".vscode", true},
		{"隐藏目录", ".idea", true},
		{"普通目录", "src", false},
		{"普通目录2", "internal", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldSkipDir(tt.dir)
			if got != tt.want {
				t.Errorf("shouldSkipDir(%q) = %v, want %v", tt.dir, got, tt.want)
			}
		})
	}
}

// TestShouldSkipFile 测试文件跳过
func TestShouldSkipFile(t *testing.T) {
	tests := []struct {
		name string
		file string
		want bool
	}{
		{"exe 文件", "test.exe", true},
		{"dll 文件", "test.dll", true},
		{"zip 文件", "test.zip", true},
		{"图片文件", "test.jpg", true},
		{"Go 文件", "test.go", false},
		{"txt 文件", "test.txt", false},
		{"md 文件", "README.md", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldSkipFile(tt.file)
			if got != tt.want {
				t.Errorf("shouldSkipFile(%q) = %v, want %v", tt.file, got, tt.want)
			}
		})
	}
}

// TestExportResults 测试导出功能
func TestExportResults(t *testing.T) {
	// 创建测试结果
	results := []SearchResult{
		{
			FilePath:  filepath.Join("test", "file1.go"),
			Lines:     []string{"func main() {", "    fmt.Println(\"hello\")"},
			LineNumbers: []int{1, 2},
			MatchType: "content",
			ContextBefore: []ContextLine{
				{LineNumber: 0, Content: "package main"},
			},
			ContextAfter: []ContextLine{
				{LineNumber: 3, Content: "}"},
			},
		},
		{
			FilePath:  filepath.Join("test", "file2.txt"),
			Lines:     []string{"file2.txt"},
			MatchType: "filename",
		},
	}

	keyword := "hello"

	t.Run("JSON 导出", func(t *testing.T) {
		outputPath := filepath.Join(t.TempDir(), "results.json")
		err := ExportResults(results, keyword, "json", outputPath)
		if err != nil {
			t.Fatalf("ExportResults JSON failed: %v", err)
		}

		// 验证文件存在且可以读取
		data, err := os.ReadFile(outputPath)
		if err != nil {
			t.Fatalf("Failed to read JSON file: %v", err)
		}

		// 验证 JSON 格式
		var result map[string]interface{}
		if err := json.Unmarshal(data, &result); err != nil {
			t.Fatalf("Invalid JSON format: %v", err)
		}

		if result["keyword"] != keyword {
			t.Errorf("keyword = %v, want %v", result["keyword"], keyword)
		}
	})

	t.Run("CSV 导出", func(t *testing.T) {
		outputPath := filepath.Join(t.TempDir(), "results.csv")
		err := ExportResults(results, keyword, "csv", outputPath)
		if err != nil {
			t.Fatalf("ExportResults CSV failed: %v", err)
		}

		// 验证文件存在且可以读取
		file, err := os.Open(outputPath)
		if err != nil {
			t.Fatalf("Failed to read CSV file: %v", err)
		}
		defer file.Close()

		reader := csv.NewReader(file)
		records, err := reader.ReadAll()
		if err != nil {
			t.Fatalf("Invalid CSV format: %v", err)
		}

		// 验证表头
		if len(records) < 1 {
			t.Fatal("CSV file is empty")
		}
		if len(records[0]) != 4 {
			t.Errorf("CSV header length = %d, want 4", len(records[0]))
		}

		// 验证数据行数（表头 + 3行数据）
		if len(records) != 4 {
			t.Errorf("CSV records = %d, want 4", len(records))
		}
	})

	t.Run("TXT 导出", func(t *testing.T) {
		outputPath := filepath.Join(t.TempDir(), "results.txt")
		err := ExportResults(results, keyword, "txt", outputPath)
		if err != nil {
			t.Fatalf("ExportResults TXT failed: %v", err)
		}

		// 验证文件存在且不为空
		info, err := os.Stat(outputPath)
		if err != nil {
			t.Fatalf("Failed to stat TXT file: %v", err)
		}
		if info.Size() == 0 {
			t.Fatal("TXT file is empty")
		}
	})

	t.Run("不支持的格式", func(t *testing.T) {
		outputPath := filepath.Join(t.TempDir(), "results.xml")
		err := ExportResults(results, keyword, "xml", outputPath)
		if err == nil {
			t.Fatal("Expected error for unsupported format, got nil")
		}
	})

	t.Run("空结果", func(t *testing.T) {
		outputPath := filepath.Join(t.TempDir(), "results.json")
		err := ExportResults([]SearchResult{}, keyword, "json", outputPath)
		if err == nil {
			t.Fatal("Expected error for empty results, got nil")
		}
	})
}

// TestMergeContexts 测试上下文合并
func TestMergeContexts(t *testing.T) {
	tests := []struct {
		name     string
		contexts [][]ContextLine
		wantLen  int
	}{
		{
			name: "无重复",
			contexts: [][]ContextLine{
				{
					{LineNumber: 1, Content: "line 1"},
					{LineNumber: 2, Content: "line 2"},
				},
				{
					{LineNumber: 5, Content: "line 5"},
					{LineNumber: 6, Content: "line 6"},
				},
			},
			wantLen: 4,
		},
		{
			name: "有重复",
			contexts: [][]ContextLine{
				{
					{LineNumber: 1, Content: "line 1"},
					{LineNumber: 2, Content: "line 2"},
				},
				{
					{LineNumber: 2, Content: "line 2"},
					{LineNumber: 3, Content: "line 3"},
				},
			},
			wantLen: 3,
		},
		{
			name:     "空上下文",
			contexts: [][]ContextLine{},
			wantLen:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			merged := mergeContexts(tt.contexts)
			if len(merged) != tt.wantLen {
				t.Errorf("mergeContexts length = %d, want %d", len(merged), tt.wantLen)
			}
		})
	}
}

// TestSearchBasic 测试基础搜索功能
func TestSearchBasic(t *testing.T) {
	// 创建临时测试目录
	testDir := t.TempDir()

	// 创建测试文件
	testFiles := map[string]string{
		"file1.go": `package main

import "fmt"

func main() {
	fmt.Println("hello world")
}`,
		"file2.txt": `This is a test file
It contains some text
for testing purposes
hello again
end of file`,
		"file3.md": `# README
This is a markdown file
No match here`,
	}

	for filename, content := range testFiles {
		filePath := filepath.Join(testDir, filename)
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
	}

	t.Run("内容搜索", func(t *testing.T) {
		config := SearchConfig{
			Keyword:   "hello",
			Path:      testDir,
			Recursive: false,
		}

		results := Search(config)

		// 应该匹配 file1.go 和 file2.txt
		if len(results) < 2 {
			t.Errorf("Expected at least 2 results, got %d", len(results))
		}

		// 验证匹配类型
		found := false
		for _, result := range results {
			if result.MatchType == "content" {
				found = true
				if len(result.Lines) == 0 {
					t.Error("Content match has no lines")
				}
			}
		}
		if !found {
			t.Error("No content match found")
		}
	})

	t.Run("文件名搜索", func(t *testing.T) {
		config := SearchConfig{
			Keyword:   "file",
			Path:      testDir,
			Recursive: false,
		}

		results := Search(config)

		// 应该匹配所有文件名
		filenameMatches := 0
		for _, result := range results {
			if result.MatchType == "filename" {
				filenameMatches++
			}
		}

		if filenameMatches < 3 {
			t.Errorf("Expected at least 3 filename matches, got %d", filenameMatches)
		}
	})

	t.Run("文件类型过滤", func(t *testing.T) {
		config := SearchConfig{
			Keyword:   "hello",
			Path:      testDir,
			Recursive: false,
			FileType:  ".go",
		}

		results := Search(config)

		// 应该只匹配 .go 文件
		for _, result := range results {
			if filepath.Ext(result.FilePath) != ".go" && result.MatchType == "content" {
				t.Errorf("Expected only .go files, got %s", result.FilePath)
			}
		}
	})

	t.Run("带上下文搜索", func(t *testing.T) {
		config := SearchConfig{
			Keyword:      "hello",
			Path:         testDir,
			Recursive:    false,
			ContextLines: 1,
		}

		results := Search(config)

		// 验证上下文是否存在
		for _, result := range results {
			if result.MatchType == "content" {
				// 可能有上下文
				if len(result.ContextBefore) > 0 || len(result.ContextAfter) > 0 {
					// 验证上下文的行号是否正确
					for _, ctx := range result.ContextBefore {
						if ctx.LineNumber <= 0 {
							t.Errorf("Invalid context before line number: %d", ctx.LineNumber)
						}
					}
					for _, ctx := range result.ContextAfter {
						if ctx.LineNumber <= 0 {
							t.Errorf("Invalid context after line number: %d", ctx.LineNumber)
						}
					}
				}
			}
		}
	})

	t.Run("空关键字", func(t *testing.T) {
		config := SearchConfig{
			Keyword:   "",
			Path:      testDir,
			Recursive: false,
		}

		results := Search(config)
		// 空关键字应该匹配所有文件
		if len(results) == 0 {
			t.Error("Expected results for empty keyword")
		}
	})

	t.Run("不存在的目录", func(t *testing.T) {
		config := SearchConfig{
			Keyword:   "hello",
			Path:      filepath.Join(testDir, "nonexistent"),
			Recursive: false,
		}

		results := Search(config)
		if len(results) != 0 {
			t.Errorf("Expected 0 results for nonexistent path, got %d", len(results))
		}
	})
}

// TestRegexSearch 测试正则表达式搜索
func TestRegexSearch(t *testing.T) {
	// 创建临时测试目录
	testDir := t.TempDir()

	// 创建测试文件
	content := `package main

import "fmt"

func main() {
	fmt.Println("hello world")
	fmt.Println("Hello World")
	fmt.Println("HELLO WORLD")
}
`
	testFile := filepath.Join(testDir, "test.go")
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	t.Run("正则表达式匹配", func(t *testing.T) {
		pattern, err := CompileRegex(`fmt\.Println\("hello`)
		if err != nil {
			t.Fatalf("Failed to compile regex: %v", err)
		}

		config := SearchConfig{
			Keyword:      `fmt\.Println\("hello`,
			Path:         testDir,
			Recursive:    false,
			UseRegex:     true,
			RegexPattern: pattern,
		}

		results := Search(config)
		if len(results) == 0 {
			t.Error("Expected results for regex search")
		}

		// 验证匹配的行
		if len(results) > 0 && len(results[0].Lines) == 0 {
			t.Error("Expected matched lines in result")
		}
	})

	t.Run("无效正则表达式", func(t *testing.T) {
		_, err := CompileRegex(`[invalid`)
		if err == nil {
			t.Error("Expected error for invalid regex pattern")
		}
	})
}

// TestMultiKeywordSearch 测试多关键字搜索
func TestMultiKeywordSearch(t *testing.T) {
	// 创建临时测试目录
	testDir := t.TempDir()

	// 创建测试文件
	content := `package main

import "fmt"

func main() {
	fmt.Println("hello world")
	fmt.Println("test function")
	fmt.Println("hello test")
}
`
	testFile := filepath.Join(testDir, "test.go")
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	t.Run("AND 模式 - 所有关键字都匹配", func(t *testing.T) {
		config := SearchConfig{
			Keyword:    "hello [AND] test",
			Path:       testDir,
			Recursive:  false,
			SearchMode: "multi_and",
			Keywords:   []string{"hello", "test"},
		}

		results := Search(config)
		if len(results) == 0 {
			t.Error("Expected results for AND mode search")
		}

		// AND 模式应该只匹配包含所有关键字的行
		if len(results) > 0 {
			found := false
			for _, line := range results[0].Lines {
				if containsIgnoreCase(line, "hello") && containsIgnoreCase(line, "test") {
					found = true
					break
				}
			}
			if !found {
				t.Error("Expected line containing both 'hello' and 'test' in AND mode")
			}
		}
	})

	t.Run("OR 模式 - 任一关键字匹配", func(t *testing.T) {
		config := SearchConfig{
			Keyword:    "hello [OR] world",
			Path:       testDir,
			Recursive:  false,
			SearchMode: "multi_or",
			Keywords:   []string{"hello", "world"},
		}

		results := Search(config)
		if len(results) == 0 {
			t.Error("Expected results for OR mode search")
		}

		// OR 模式应该匹配包含任一关键字的行
		if len(results) > 0 {
			for _, line := range results[0].Lines {
				if !containsIgnoreCase(line, "hello") && !containsIgnoreCase(line, "world") {
					t.Errorf("Expected line containing 'hello' or 'world', got: %s", line)
				}
			}
		}
	})

	t.Run("空关键字列表", func(t *testing.T) {
		result := matchContentAnd("test", []string{})
		if result {
			t.Error("Expected false for empty keywords in AND mode")
		}

		result = matchContentOr("test", []string{})
		if result {
			t.Error("Expected false for empty keywords in OR mode")
		}
	})
}

// TestIgnorePatterns 测试忽略文件功能
func TestIgnorePatterns(t *testing.T) {
	// 创建临时测试目录
	testDir := t.TempDir()

	// 创建测试文件
	files := map[string]string{
		"test1.go":    "package main",
		"test2.go":    "package main",
		"test.txt":    "text content",
		"ignore.go":   "package main",
		".hidden.txt": "hidden file",
	}

	for name, content := range files {
		filePath := filepath.Join(testDir, name)
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", name, err)
		}
	}

	// 创建 .gitignore 文件
	gitignoreContent := `ignore.go
*.txt
`
	gitignorePath := filepath.Join(testDir, ".gitignore")
	if err := os.WriteFile(gitignorePath, []byte(gitignoreContent), 0644); err != nil {
		t.Fatalf("Failed to create .gitignore: %v", err)
	}

	t.Run("加载 .gitignore", func(t *testing.T) {
		patterns := loadIgnorePatterns(testDir)
		if len(patterns) == 0 {
			t.Error("Expected patterns from .gitignore")
		}

		// 验证模式是否正确加载
		found := false
		for _, p := range patterns {
			if p == "ignore.go" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected 'ignore.go' pattern in loaded patterns")
		}
	})

	t.Run("shouldIgnorePath 测试", func(t *testing.T) {
		patterns := []string{"ignore.go", "*.txt"}

		if !shouldIgnorePath("ignore.go", patterns) {
			t.Error("Expected 'ignore.go' to be ignored")
		}

		if !shouldIgnorePath("test.txt", patterns) {
			t.Error("Expected '*.txt' to match 'test.txt'")
		}

		if shouldIgnorePath("test.go", patterns) {
			t.Error("Did not expect 'test.go' to be ignored")
		}
	})

	t.Run("搜索时应用忽略模式", func(t *testing.T) {
		config := SearchConfig{
			Keyword:       "package",
			Path:          testDir,
			Recursive:     false,
			LoadGitignore: true,
		}

		results := Search(config)

		// 验证 ignore.go 和 test.txt 被忽略
		for _, result := range results {
			if filepath.Base(result.FilePath) == "ignore.go" {
				t.Error("Expected 'ignore.go' to be ignored in search results")
			}
			if filepath.Base(result.FilePath) == "test.txt" {
				t.Error("Expected '*.txt' files to be ignored in search results")
			}
		}
	})
}

// TestMatchIgnorePattern 测试忽略模式匹配
func TestMatchIgnorePattern(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		pattern string
		want    bool
	}{
		{"精确匹配", "test.go", "test.go", true},
		{"后缀匹配", "src/test.go", "test.go", true},
		{"前缀匹配", "test.min.js", "*.min.js", true},
		{"不匹配", "test.go", "ignore.go", false},
		{"目录后缀匹配", "src/build", "build/", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchIgnorePattern(tt.path, tt.pattern)
			if got != tt.want {
				t.Errorf("matchIgnorePattern(%q, %q) = %v, want %v", tt.path, tt.pattern, got, tt.want)
			}
		})
	}
}

