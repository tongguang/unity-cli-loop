# Tuanjie CLI Loop (`tloop`)

简体中文 | [日本語](README_ja.md) | [Upstream (English)](https://github.com/hatayama/unity-cli-loop)

[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE.md)

让 AI 通过命令行编译、测试和操作团结引擎项目。

本项目基于 **Masamichi Hatayama** 的 [Unity CLI Loop (`uloop`)](https://github.com/hatayama/unity-cli-loop)，由 [tongguang](https://github.com/tongguang) 维护团结专用适配。上游的开发循环、工具实现和 skills 是本项目的基础；本分支主要调整团结 Editor 的发现、运行环境隔离和发布流程。

## 分支定位

`tloop` 面向团结引擎。需要操作原版 Unity 工程时，请使用上游 `uloop`。

两者可以安装在同一台机器上，各自使用独立的命令、缓存、环境变量、项目状态和 skills：

| 项目 | 团结分支 |
| --- | --- |
| CLI 命令 | `tloop` |
| Project runner | `tloop-project-runner` |
| UPM 包名 | `io.github.tongguang.tloop` |
| 环境变量 | `TLOOP_*` |
| 项目状态和输出 | `.tloop/` |
| 安装的 skills | `tloop-*` |

CLI 与 Editor 使用本地 IPC：Windows 使用命名管道，macOS / Linux 使用 Unix domain socket。端点按工程路径区分，不需要设置起始端口。

团结仍提供 `UnityEngine`、`UnityEditor` 等 API。本分支保留既有程序集和扩展 API 的名称，便于已有自定义工具继续使用。

## 版本与验证范围

三个组件分别发布版本，project runner 由工程内的 pin 自动选择：

| 组件 | 当前版本 |
| --- | --- |
| 团结 UPM 插件 | [3.22.1](https://github.com/tongguang/unity-cli-loop/releases/tag/v3.22.1) |
| CLI dispatcher | [3.14.1](https://github.com/tongguang/unity-cli-loop/releases/tag/dispatcher-v3.14.1) |
| Project runner | [3.13.1](https://github.com/tongguang/unity-cli-loop/releases/tag/tloop-project-runner-v3.13.1) |

已在 Windows 上使用 **团结 1.10.3 / `2022.3.62t15`** 完成无界面导入、IPC 连通、90 项定向测试和 57 个程序集的离线编译检查。

图形界面启动流程、PlayMode、热重载和录屏等功能仍需进一步验证；其他团结版本以及 macOS / Linux 上的 Editor 适配也尚未实测。已有 Unity CI 不能作为团结兼容性的证明。详细结果与已知限制见 [团结兼容性说明](docs/tuanjie-compatibility.md)。

## 安装

需要已安装的团结 Editor，以及可加载 skills 的 AI 工具。

### 1. 安装 CLI

Windows PowerShell：

```powershell
irm https://raw.githubusercontent.com/tongguang/unity-cli-loop/main/scripts/install.ps1 | iex
```

macOS / Linux / Windows Git Bash：

```sh
curl -fsSL https://raw.githubusercontent.com/tongguang/unity-cli-loop/main/scripts/install.sh | sh
```

检查安装结果：

```sh
tloop --version
```

本分支暂未配置 Homebrew、winget 或 OpenUPM 分发。`tloop package install` 依赖 OpenUPM 注册；本次发布请按下一步使用 Git URL 或 tarball 安装插件。

### 2. 安装团结插件

在团结 Editor 的 `Window > Package Manager` 中选择 **Add package from git URL**，输入：

```text
https://github.com/tongguang/unity-cli-loop.git?path=/Packages/src#v3.22.1
```

也可以在已有 `Packages/manifest.json` 的 `dependencies` 中添加这一项，保留工程原有依赖：

```json
{
  "dependencies": {
    "io.github.tongguang.tloop": "https://github.com/tongguang/unity-cli-loop.git?path=/Packages/src#v3.22.1"
  }
}
```

另一种方式是下载 [UPM 发行包](https://github.com/tongguang/unity-cli-loop/releases/download/v3.22.1/io.github.tongguang.tloop-3.22.1.tgz)，选择 **Add package from tarball**。

插件按普通 UPM 包发布，无需 Unity Cloud 包签名。CLI 发行包继续使用校验和及来源证明校验。

### 3. 安装 skills

插件本身包含 `tloop-*` skill 源文件。安装插件后，在团结工程根目录运行：

```sh
# Claude Code
tloop skills install --claude

# Codex and other tools that read .agents/skills
tloop skills install --agents
```

可选安装位置：

```sh
# Install at the containing Git repository root
tloop skills install --claude --agents --root git

# Install globally
tloop skills install --claude --global
```

也可以在 Editor 的 `Window > Tuanjie CLI Loop > Settings` 中使用 **Install CLI** 和 **Install Skills**。`tloop-*` skills 与已有的 `uloop-*` skills 使用不同名称。

### 4. 连接工程

先用团结打开并保存工程，让 `ProjectSettings/ProjectVersion.txt` 记录团结版本。随后在该工程根目录运行：

```sh
tloop status
tloop list
```

Editor 未运行时，可使用 `tloop launch`。它会从工程记录的 `m_EditorVersion` 自动查找团结 Hub 注册的安装和默认安装目录。自定义安装也可以显式指定可执行文件：

```powershell
tloop launch --editor-path 'D:\Engines\Tuanjie\Editor\Tuanjie.exe'
```

自动查找以 `2022.3.62t15` 这类 Editor 版本为准；`1.10.3` 是团结产品版本。具体规则见 [Editor 发现与本地验证](docs/tuanjie-compatibility.md#editor-discovery-and-local-validation)。

## 常用操作

在目标团结工程根目录执行：

```sh
tloop compile
tloop get-logs --max-count 10
tloop run-tests --filter-type class --filter-value YourTestClass
tloop get-hierarchy
tloop screenshot
tloop execute-dynamic-code --code 'using UnityEngine; Debug.Log("Hello from tloop!");'
```

| 场景 | 命令 |
| --- | --- |
| 状态、取消任务 | `status`、`cancel` |
| 编译、离线编译检查 | `compile`、`compile-check` |
| 运行测试、检查 Console | `run-tests`、`get-logs`、`clear-console` |
| 热重载、源码暂停点 | `hot-reload`、`pause-point` |
| 场景和对象检查 | `get-hierarchy`、`find-game-objects` |
| 动态执行 C# | `execute-dynamic-code` |
| 窗口、截图和录屏 | `focus-window`、`screenshot`、`record-video` |
| Game View 和设备模拟 | `set-game-view-size`、`game-view-mode` |
| PlayMode 与输入回放 | `control-play-mode`、`simulate-mouse-ui`、`simulate-mouse-input`、`simulate-keyboard`、`replay-input` |

此表列出已有工具，具体的团结实测范围见上文。输入模拟依赖工程安装 Input System；Linux 不支持 `focus-window`。命令参数可用 `tloop <command> --help` 查看。

## 扩展与参考

- [自定义工具开发](Packages/src/Documentation~/custom-tools.md)：扩展 C# 工具及对应的 `Skill/SKILL.md`。
- [Project runner pin](docs/project-runner-pin.md)：dispatcher 如何选择工程对应的 runner 版本。
- [团结兼容性与运行环境隔离](docs/tuanjie-compatibility.md)。
- [普通 UPM 包的发布流程](docs/unity-package-release.md)。
- [Claude Code 沙箱与 IPC](docs/claude-code-sandbox.md)：连接被沙箱拒绝时的排查说明。
- [上游功能演示](https://github.com/hatayama/unity-cli-loop#demos)：由上游作者提供，展示原版 Unity / `uloop` 的使用，不代表团结验证结果。仓库内沿用的 logo、截图和演示素材也来自上游；上游 logo 注明灵感来自 Daft Punk 的 *Human After All* 专辑封面。

## 开源许可与致谢

本项目沿用上游的 **MIT License**，原始版权声明为：

> Copyright (c) 2025 Masamichi Hatayama

完整版权、许可条件和免责声明保留在 [LICENSE.md](LICENSE.md)。本分支的更名和团结适配不改变上游代码的作者归属。复制、修改或再分发本项目时，请保留该版权声明及完整许可文本。

包内第三方组件另见 [THIRD-PARTY-NOTICES.md](Packages/src/THIRD-PARTY-NOTICES.md) 及相应插件目录的许可证。这些声明继续随插件分发。
