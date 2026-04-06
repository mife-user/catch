@echo off
REM Catch Windows 基准测试脚本
REM 用法: run_benchmarks.bat [输出文件]

setlocal

set OUTPUT_FILE=%1
if "%OUTPUT_FILE%"=="" set OUTPUT_FILE=benchmark_results.txt

echo.
echo 🚀 开始运行 Catch 基准测试...
echo 输出文件: %OUTPUT_FILE%
echo.

REM 运行基准测试
echo ================================== > "%OUTPUT_FILE%"
echo Catch 基准测试结果 >> "%OUTPUT_FILE%"
echo 日期: %date% %time% >> "%OUTPUT_FILE%"
echo ================================== >> "%OUTPUT_FILE%"
echo. >> "%OUTPUT_FILE%"

echo 📊 正在执行基准测试...
go test -bench=. -benchmem -run=^$ ./internal/search -count=3 >> "%OUTPUT_FILE%" 2>&1

echo. >> "%OUTPUT_FILE%"
echo ================================== >> "%OUTPUT_FILE%"
echo 测试完成 >> "%OUTPUT_FILE%"
echo ================================== >> "%OUTPUT_FILE%"

echo.
echo ✅ 基准测试完成！
echo 📄 结果已保存到: %OUTPUT_FILE%
echo.
echo 查看结果:
echo   type %OUTPUT_FILE%
echo.

pause
