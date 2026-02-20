# Windows WIMGAPI 的 Golang 绑定规划

## 1. 目标与范围
- 目标：构建一个可维护、可测试的 Go 包，以类型安全、Go 风格 API 封装 `wimgapi.dll`。
- 目标平台：仅 Windows（优先 `windows/amd64`，后续支持 `windows/arm64`）。
- MVP 能力：
  - 打开和关闭 WIM 文件。
  - 列出映像并读取基础元数据（index、name、description、flags）。
  - 加载映像并应用到目标目录。
  - 提供基础进度回调与错误映射。

## 2. v0 不包含的范围
- 覆盖全部 WIMGAPI 函数。
- 跨平台仿真。
- GUI 工具。
- 完整 DISM 集成（仅保留扩展点）。

## 3. 绑定策略
- 主方案：使用 `golang.org/x/sys/windows` 动态加载：
  - `windows.NewLazySystemDLL("wimgapi.dll")`
  - `NewProc(...)`
- 强约束：任意包、构建标签、测试路径都不允许 cgo。
- 原则：仅纯 Go（`syscall`/`x/sys/windows`），保证构建简单、Windows Runner 可移植。

## 4. 包结构
建议模块路径：`github.com/<org>/go-wimgapi`

```text
/wimgapi
  errors.go           // Win32/HRESULT 到 Go error
  types.go            // 常量、枚举、结构体、句柄包装
  dll.go              // 延迟加载 dll + proc 绑定 + 版本检查
  file.go             // WIMCreateFile/WIMCloseHandle 封装
  image.go            // WIMLoadImage/WIMGetImageInformation 封装
  apply.go            // WIMApplyImage 封装 + 选项
  callback.go         // 回调注册与事件分发
  mount.go            // mount/unmount（第 2 阶段）
  capture.go          // capture/create（第 2 阶段）
/internal/syscallx
  utf16.go
  handle.go
  unsafeconv.go
/cmd/wimctl
  main.go             // 用于集成验证的最小 CLI
```

## 5. 对外 API 形态（Go 优先）
- `Open(path string, opts OpenOptions) (*File, error)`
  - 封装 `WIMCreateFile`。
- `(*File) Close() error`
- `(*File) Images() ([]ImageInfo, error)`
  - 组合计数、加载、元数据解析。
- `(*File) LoadImage(index int) (*Image, error)`
- `(*Image) Apply(target string, opts ApplyOptions) error`
  - 封装 `WIMApplyImage`。
- 回调模型：
  - `type ProgressFunc func(evt ProgressEvent) (cancel bool)`
  - 将 WIMGAPI 消息转换为类型化 Go 事件。

## 6. 错误与资源模型
- 所有导出资源类型必须显式 `Close()`。
- 可选 finalizer 仅用于泄漏告警，不用于正常清理。
- 统一错误类型：
  - `type Error struct { Op string; Code uint32; Msg string }`
- 通过 `FormatMessage` 补充 Win32 可读错误文本。
- 为常见错误提供判断辅助：
  - 权限不足
  - 路径无效
  - 映像索引越界

## 7. 回调与并发规则
- 原生回调必须尽量轻量且非阻塞。
- 通过内部 channel 将回调事件封送到 Go。
- 明确线程安全边界：
  - `File` 与 `Image` 默认不保证并发安全。
- 支持通过回调返回值取消，并在可行时支持 context。
- 回调桥接仅允许 `syscall.NewCallback`（禁止 cgo trampoline）。

## 8. 构建与兼容性
- 所有实现文件使用 `//go:build windows`。
- 运行时检测 `wimgapi.dll` 缺失并返回可诊断错误。
- 文档需说明环境依赖：
  - ADK / OS 版本差异
  - mount/apply 场景所需管理员权限
- CI 强制 no-cgo：
  - 使用 `CGO_ENABLED=0` 构建/测试
  - 增加检查，发现 `import "C"` 即失败

## 9. 测试计划
- 单元测试：
  - 选项翻译
  - UTF-16 转换
  - 错误转换
  - 入参校验
- 集成测试（Windows Runner）：
  - 基于小型 fixture WIM 覆盖 open/list/load/apply。
- 可选 E2E：
  - 对临时目标目录执行 `cmd/wimctl` 并校验输出目录树。
- CI 建议：
  - GitHub Actions `windows-latest`
  - 分离 `unit` 与 `integration` 任务。

## 10. 里程碑
- M0（1-2 天）：脚手架
  - 模块初始化、目录结构、dll loader、基础错误。
- M1（3-5 天）：文件与映像读取链路
  - Open/Close/Images/LoadImage + 测试。
- M2（3-5 天）：apply 支持
  - Apply + 回调 + CLI 演示。
- M3（4-7 天，可选）：mount/capture
  - mount、commit、unmount，并补权限约束文档。
- M4（2-3 天）：发布准备
  - README、示例、CI 打磨，打 `v0.1.0` 标签。

## 11. 风险与缓解
- 风险：不同 Windows 版本行为差异。
  - 缓解：建立 Win10/Win11/Server 集成矩阵。
- 风险：回调边界不稳定。
  - 缓解：最小化回调逻辑、严格内存归属规则、补充回调生命周期压力测试。
- 风险：高权限测试不稳定。
  - 缓解：隔离高权限用例，不阻塞核心单元测试。

## 12. 文档与发布规范
- README 必须包含：
  - 支持范围
  - 安装与快速开始
  - 权限要求
  - 已知限制
- 增加 `examples/`：
  - `list-images`
  - `apply-image`
- 版本策略：
  - 在 `v0.x` 快速迭代，保持 API 变更日志清晰。

## 13. 首轮迭代任务清单
1. 初始化 Go 模块与目录骨架。
2. 实现 `dll.go` 延迟 proc 绑定。
3. 使用 `syscall.NewCallback` 实现回调管线并验证 no-cgo 构建。
4. 实现 `errors.go`（Win32 消息转换）。
5. 实现 `file.go`（`Open` 与 `Close`）。
6. 实现 `image.go`（最小 list/load 流程）。
7. 增加 `cmd/wimctl list <wim>` 做冒烟验证。
8. 增加单元测试与基础 Windows CI 工作流。
