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
