# go-wimgapi

`wimgapi.dll` 的纯 Go（无 cgo）Windows 绑定。

##  状态

- 初始实现（v0 范围）。
- 平台：仅支持 Windows。
- cgo：不使用；CI/构建应使用 `CGO_ENABLED=0`。

## 已实现 API

- 打开/关闭 WIM 文件
- 获取映像数量与映像元数据列表
- 按索引加载映像
- 将映像应用到目标目录
- 为 apply 注册进度回调
- 为 capture 注册进度回调
- 使用 `ProgressDecoder` 解码回调消息

## CLI

`cmd/wimctl` 提供：

- `wimctl list <path-to-wim>`
- `wimctl apply <path-to-wim> <index> <target-dir>`
- `wimctl capture <source-dir> <path-to-wim>`

## 快速开始

```powershell
go run ./cmd/wimctl list C:\images\install.wim
```

## Roundtrip 测试程序

```powershell
go run ./examples/roundtrip-test
```

集成测试包装（可选）：

```powershell
$env:WIMGAPI_INTEGRATION='1'
go test -run TestRoundtripProgram ./...
```

## 说明

- 部分操作可能需要管理员权限。
- 运行时行为可能随 Windows/ADK 版本变化。
- 如果出现临时目录错误（例如错误码 `1632`），请设置可写的 WIM 临时路径（roundtrip 示例已自动处理）。
