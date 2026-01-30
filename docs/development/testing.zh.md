# 测试指南

测试 clinvk 的指南和实践。

## 运行测试

### 所有测试

```bash
go test ./...
```

### 带竞态检测

```bash
go test -race ./...
```

### 带覆盖率

```bash
go test -coverprofile=coverage.txt ./...
go tool cover -html=coverage.txt
```

### 简短测试

```bash
go test -short ./...
```

### 详细输出

```bash
go test -v ./...
```

### 使用 Just

```bash
just test           # 运行所有测试
just test-verbose   # 详细输出
just test-short     # 仅简短测试
just test-coverage  # 生成覆盖率
```

## 编写测试

### 表驱动测试

使用表驱动测试处理多个场景：

```go
func TestParseOutput(t *testing.T) {
    t.Parallel()

    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {"简单输出", "Hello, world!", "Hello, world!", false},
        {"带空白", "  trimmed  ", "trimmed", false},
        {"空输入", "", "", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()
            got, err := ParseOutput(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("got %q, want %q", got, tt.want)
            }
        })
    }
}
```

## Mock 包

`internal/mock` 包提供测试工具。

### Mock 后端

```go
import "github.com/signalridge/clinvoker/internal/mock"

func TestWithMockBackend(t *testing.T) {
    mb := mock.NewMockBackend("test",
        mock.WithParseOutput("mocked output"),
        mock.WithAvailable(true),
    )

    output := mb.ParseOutput("any input")
    if output != "mocked output" {
        t.Errorf("expected mocked output")
    }
}
```

## 基准测试

```go
func BenchmarkParseOutput(b *testing.B) {
    input := "large output content..."

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        ParseOutput(input)
    }
}
```

运行基准测试：

```bash
go test -bench=. ./...
```

## 覆盖率目标

- 核心包覆盖率 >80%
- 专注于行为，而非行覆盖率
- 不跳过错误路径

## CI 集成

测试自动运行于：

- Pull requests
- 推送到 main
- 发布标签

详见 `.github/workflows/ci.yaml`。
