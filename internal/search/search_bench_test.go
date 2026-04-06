package search

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// createTestDir 创建测试目录并生成测试文件
func createTestDir(b *testing.B, numFiles int, linesPerFile int) string {
	testDir := b.TempDir()

	for i := 0; i < numFiles; i++ {
		filename := fmt.Sprintf("testfile_%d.go", i)
		filePath := filepath.Join(testDir, filename)

		content := generateTestContent(linesPerFile, i)
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			b.Fatalf("创建测试文件失败: %v", err)
		}
	}

	return testDir
}

// generateTestContent 生成测试内容
func generateTestContent(lines int, fileIndex int) string {
	content := fmt.Sprintf("package test%d\n\n", fileIndex)
	content += "import \"fmt\"\n\n"

	for i := 0; i < lines; i++ {
		if i%10 == 0 {
			content += fmt.Sprintf("// Function hello%d - this is a test function\n", i)
			content += fmt.Sprintf("func hello%d() {\n", i)
			content += fmt.Sprintf("    fmt.Println(\"hello world %d\")\n", i)
			content += "}\n\n"
		} else {
			content += fmt.Sprintf("// Line %d: some comment about testing and hello\n", i)
			content += fmt.Sprintf("var variable%d = %d\n", i, i)
		}
	}

	return content
}

// BenchmarkSearch_SmallDir 基准测试：小目录（10 个文件）
func BenchmarkSearch_SmallDir(b *testing.B) {
	testDir := createTestDir(b, 10, 50)

	config := SearchConfig{
		Keyword:      "hello",
		Path:         testDir,
		Recursive:    false,
		MaxGoroutine: 5,
		Context:      context.Background(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Search(config)
	}
}

// BenchmarkSearch_MediumDir 基准测试：中等目录（100 个文件）
func BenchmarkSearch_MediumDir(b *testing.B) {
	testDir := createTestDir(b, 100, 50)

	config := SearchConfig{
		Keyword:      "hello",
		Path:         testDir,
		Recursive:    false,
		MaxGoroutine: 10,
		Context:      context.Background(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Search(config)
	}
}

// BenchmarkSearch_LargeDir 基准测试：大目录（500 个文件）
func BenchmarkSearch_LargeDir(b *testing.B) {
	testDir := createTestDir(b, 500, 50)

	config := SearchConfig{
		Keyword:      "hello",
		Path:         testDir,
		Recursive:    false,
		MaxGoroutine: 10,
		Context:      context.Background(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Search(config)
	}
}

// BenchmarkSearch_Recursive 基准测试：递归搜索
func BenchmarkSearch_Recursive(b *testing.B) {
	testDir := createTestDir(b, 100, 50)

	// 创建子目录
	subDir := filepath.Join(testDir, "subdir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		b.Fatalf("创建子目录失败: %v", err)
	}

	// 在子目录创建文件
	for i := 0; i < 50; i++ {
		filename := fmt.Sprintf("subfile_%d.go", i)
		filePath := filepath.Join(subDir, filename)
		content := generateTestContent(50, i+100)
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			b.Fatalf("创建测试文件失败: %v", err)
		}
	}

	config := SearchConfig{
		Keyword:      "hello",
		Path:         testDir,
		Recursive:    true,
		MaxGoroutine: 10,
		Context:      context.Background(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Search(config)
	}
}

// BenchmarkSearchWithContext 基准测试：带上下文搜索
func BenchmarkSearchWithContext(b *testing.B) {
	testDir := createTestDir(b, 100, 100)

	contextSizes := []int{0, 1, 2, 5}

	for _, ctxLines := range contextSizes {
		b.Run(fmt.Sprintf("ContextLines=%d", ctxLines), func(b *testing.B) {
			config := SearchConfig{
				Keyword:      "hello",
				Path:         testDir,
				Recursive:    false,
				MaxGoroutine: 10,
				ContextLines: ctxLines,
				Context:      context.Background(),
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				Search(config)
			}
		})
	}
}

// BenchmarkPageResults_Small 基准测试：小规模分页
func BenchmarkPageResults_Small(b *testing.B) {
	results := make([]SearchResult, 100)
	for i := 0; i < 100; i++ {
		results[i] = SearchResult{
			FilePath:  filepath.Join("test", fmt.Sprintf("file%d.go", i)),
			Lines:     []string{"test line"},
			MatchType: "content",
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		PageResults(results, 1, 10)
	}
}

// BenchmarkPageResults_Large 基准测试：大规模分页
func BenchmarkPageResults_Large(b *testing.B) {
	results := make([]SearchResult, 10000)
	for i := 0; i < 10000; i++ {
		results[i] = SearchResult{
			FilePath:  filepath.Join("test", fmt.Sprintf("file%d.go", i)),
			Lines:     []string{"test line"},
			MatchType: "content",
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		PageResults(results, 50, 10)
	}
}

// BenchmarkContainsIgnoreCase 基准测试：大小写无关匹配
func BenchmarkContainsIgnoreCase(b *testing.B) {
	tests := []struct {
		name string
		s    string
		sub  string
	}{
		{"短字符串", "hello world", "world"},
		{"长字符串", "this is a very long string with hello world in the middle", "HELLO"},
		{"不匹配", "abcdefghijklmnopqrstuvwxyz", "xyz123"},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				containsIgnoreCase(tt.s, tt.sub)
			}
		})
	}
}

// BenchmarkExportJSON 基准测试：导出 JSON
func BenchmarkExportJSON(b *testing.B) {
	results := make([]SearchResult, 100)
	for i := 0; i < 100; i++ {
		results[i] = SearchResult{
			FilePath:    filepath.Join("test", fmt.Sprintf("file%d.go", i)),
			Lines:       []string{fmt.Sprintf("line %d with hello", i)},
			LineNumbers: []int{i + 1},
			MatchType:   "content",
			ContextBefore: []ContextLine{
				{LineNumber: i, Content: fmt.Sprintf("context before %d", i)},
			},
			ContextAfter: []ContextLine{
				{LineNumber: i + 2, Content: fmt.Sprintf("context after %d", i)},
			},
		}
	}

	outputPath := filepath.Join(b.TempDir(), "benchmark_results.json")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ExportResults(results, "hello", "json", outputPath)
	}
}

// BenchmarkExportCSV 基准测试：导出 CSV
func BenchmarkExportCSV(b *testing.B) {
	results := make([]SearchResult, 100)
	for i := 0; i < 100; i++ {
		results[i] = SearchResult{
			FilePath:    filepath.Join("test", fmt.Sprintf("file%d.go", i)),
			Lines:       []string{fmt.Sprintf("line %d with hello", i)},
			LineNumbers: []int{i + 1},
			MatchType:   "content",
		}
	}

	outputPath := filepath.Join(b.TempDir(), "benchmark_results.csv")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ExportResults(results, "hello", "csv", outputPath)
	}
}

// BenchmarkExportTXT 基准测试：导出 TXT
func BenchmarkExportTXT(b *testing.B) {
	results := make([]SearchResult, 100)
	for i := 0; i < 100; i++ {
		results[i] = SearchResult{
			FilePath:    filepath.Join("test", fmt.Sprintf("file%d.go", i)),
			Lines:       []string{fmt.Sprintf("line %d with hello", i)},
			LineNumbers: []int{i + 1},
			MatchType:   "content",
		}
	}

	outputPath := filepath.Join(b.TempDir(), "benchmark_results.txt")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ExportResults(results, "hello", "txt", outputPath)
	}
}
