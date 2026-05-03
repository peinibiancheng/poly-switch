我正在使用 Golang 开发一个名为 poly-switch 的多语言版本管理器。它的设计灵感来自 cc-switch-cli。
技术栈：
UI 框架：使用 charmbracelet/bubbletea 作为核心。
组件库：使用 bubbles 处理列表选择，使用 lipgloss 进行样式设计。
持久化：使用 JSON 或 YAML 存储版本信息和当前激活状态。
核心逻辑：
劫持思想：工具不直接修改系统环境，而是通过维护一个 ~/.poly-switch/bin 目录，在其中创建软链接（Symlinks）。用户只需将该目录加入系统 PATH。
插件化配置：支持 Java, Python, Node。每种语言有对应的 Home 路径定义和版本解析逻辑。
交互流程：运行 poly-switch -> 显示支持的语言列表 -> 选择语言 -> 显示已安装版本列表 -> 选定版本 -> 自动更新软链接并同步状态。
请帮我完成以下任务：
设计一个能够描述多种语言及其版本的 Config 结构体。
编写 Bubble Tea 的基础模型（Model），包含“语言列表选择”和“版本列表选择”两个阶段的状态机。
实现一个核心函数 applySwitch(lang, version)，用于根据配置更新文件系统中的软链接。
提供符合 Lipgloss 风格的极简、美观的 UI 代码片段。
请确保代码结构清晰，逻辑解耦，方便后续扩展更多编程语言。


软链接是关键：在用户的 ~/.zshrc 或 ~/.bashrc 中，只需要一行：
export PATH="$HOME/.poly-switch/bin:$PATH"
你的 Go 程序只需要把 ~/.poly-switch/bin/java 指向具体的 JDK 路径，切换瞬间完成，无需重启终端。
并发检查：利用 Go 的 goroutine，在 Bubble Tea 启动的 Init() 阶段，可以异步扫描本地已安装的 Java/Python 路径。
状态反馈：Bubble Tea 的 Spinner 组件非常适合在 applySwitch 期间显示“正在切换...”的动效。

### UI 设计补充：镜像管理与状态交互

1. **状态降级（视觉补位）**
   对于没有镜像源概念的语言，不要直接留白（会显得像 Bug），而是显示一个“不可用”或“系统接管”的状态。
   - 有镜像源（如 Node）：`Mirror:  🚀 npmmirror.com (12ms)`
   - 无镜像源（如部分纯本地语言）：`Mirror:  🔒 System Default (N/A)`  — 使用灰色文字，提示该语言目前不通过 poly-switch 管理镜像。

2. **动态 Context 布局**
   右侧详情栏不应该是“死板”的固定字段，而是根据左侧选中的语言动态渲染不同的组件卡片。
   - 当选中 Node 时：右侧渲染 `[状态] [版本] [镜像源]` 等区块。
   - 当选中 Java 时：右侧除了 `[状态] [版本]`，可能还会渲染 `[JDK 类型]`（比如 OpenJDK/Oracle）。
   - 实现思路：在 Bubbletea 的 View 函数中，根据 `m.selectedLanguage` 使用 switch 语句分发不同的渲染逻辑。

3. **底部按键的“灰度处理”**
   这是交互体验中最重要的细节：如果当前语言没有某些管理功能，底部的快捷键提示应该改变。
   - 选中支持镜像的语言：底部显示 `[M] Manage Mirrors`（正常/高亮）。
   - 选中不支持镜像的语言：底部显示 `[M] Manage Mirrors`（暗灰色/删除线）或者直接隐藏该按键提示。
