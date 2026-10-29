# Tuanjie CLI Loop (`tloop`)

简体中文 | [日本語](README_ja.md) | [Upstream (English)](https://github.com/hatayama/unity-cli-loop)

[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE.md)

`tloop` 是一个命令行工具，让 Claude Code、Codex 等 AI 编程助手直接操作团结 Editor：编译、跑测试、读 Console 日志、截图、执行 C#。

AI 改完代码可以自己编译，看到报错就接着修，直到编译通过、测试变绿。你不用再把报错从 Editor 复制到聊天窗口。

本项目是 Masamichi Hatayama 的 [Unity CLI Loop](https://github.com/hatayama/unity-cli-loop)（`uloop`）的团结版，由 [tongguang](https://github.com/tongguang) 维护。原版 Unity 工程请用上游的 `uloop`，两者装在同一台机器上互不影响。

## 能做什么

- **编译和测试**：编译工程并返回错误和警告，运行 EditMode / PlayMode 测试，读取和清空 Console。
- **快速验证改动**：热重载把 C# 修改直接应用到正在运行的 Editor，不触发域重载，PlayMode 状态也会保留。离线编译检查不连接 Editor 也能查出编译错误。
- **查看场景**：输出场景层级，查找 GameObject，截取 Editor 窗口或 Game View，录制视频。
- **调试运行时行为**：控制 PlayMode，模拟键盘和鼠标输入，在指定源码行暂停并读取局部变量，或统计这行代码执行了几次。
- **执行任意 C#**：现有命令覆盖不到的操作，可以让 AI 直接在 Editor 里执行一段 C#。
- **扩展**：用 C# 编写自己的命令，再配一份给 AI 看的 skill。

每个命令都配有一份 skill，告诉 AI 这个命令什么时候用、怎么用。装好 skills 后，你用平常的话提需求就行，AI 会自己选命令。

## 快速开始

需要：团结 Editor，以及支持 skills 的 AI 工具（如 Claude Code、Codex）。

**1. 安装 CLI**

Windows（PowerShell）：

```powershell
irm https://raw.githubusercontent.com/tongguang/unity-cli-loop/main/scripts/install.ps1 | iex
```

macOS / Linux / Windows Git Bash：

```sh
curl -fsSL https://raw.githubusercontent.com/tongguang/unity-cli-loop/main/scripts/install.sh | sh
```

运行 `tloop --version` 确认安装成功。提示找不到命令时，重新打开终端。

**2. 把插件加入工程**

在团结工程根目录运行：

```sh
tloop package install
```

这条命令会把插件写进 `Packages/manifest.json`。然后用团结打开工程（已经打开的话切回 Editor），等插件导入完成。

**3. 安装 skills**

skills 随插件一起分发，所以要等插件导入后再装。仍在工程根目录运行：

```sh
# Claude Code
tloop skills install --claude

# Codex 以及其他读取 .agents/skills 的工具
tloop skills install --agents
```

**4. 试一下**

```sh
tloop status
```

输出里出现 `"State": "Ready"`，说明 CLI 已经连上 Editor。现在可以在 AI 工具里说：

> 用 tloop 编译这个工程，有报错就修好。

## 常用命令

这些命令平时由 AI 调用，你在终端里手动运行也一样：

```sh
tloop compile
tloop get-logs --max-count 10
tloop run-tests --filter-type class --filter-value YourTestClass
tloop screenshot
tloop execute-dynamic-code --code 'using UnityEngine; Debug.Log("Hello from tloop!");'
```

| 用途 | 命令 |
| --- | --- |
| 查看 Editor 状态、取消正在执行的命令 | `status`、`cancel` |
| 启动或重启 Editor | `launch` |
| 编译、离线编译检查 | `compile`、`compile-check` |
| 测试和 Console | `run-tests`、`get-logs`、`clear-console` |
| 热重载 | `hot-reload` |
| 场景和对象 | `get-hierarchy`、`find-game-objects` |
| 执行任意 C# | `execute-dynamic-code` |
| 窗口、截图、录屏 | `focus-window`、`screenshot`、`record-video` |
| Game View 和设备模拟器 | `set-game-view-size`、`game-view-mode` |
| PlayMode 和输入 | `control-play-mode`、`simulate-keyboard`、`simulate-mouse-input`、`simulate-mouse-ui`、`replay-input` |
| 源码暂停点和变量监视 | `enable-pause-point`、`await-pause-point`、`enable-watch`、`get-watch-values` |

`tloop list` 列出当前工程可用的全部命令，`tloop <command> --help` 查看参数。输入模拟需要工程使用 Input System 包。`focus-window` 不支持 Linux。

## 其他安装方式

**固定插件版本**：`tloop package install` 默认跟随 `main` 分支。要固定到某个发布版本，运行 `tloop package install --version <版本号>`。版本号见 [Releases](https://github.com/tongguang/unity-cli-loop/releases)。

**在 Package Manager 里添加**：在 `Window > Package Manager` 中选择 **Add package from git URL**，输入：

```text
https://github.com/tongguang/unity-cli-loop.git?path=/Packages/src
```

也可以手动在 `Packages/manifest.json` 的 `dependencies` 里加上这一项：

```json
"io.github.tongguang.tloop": "https://github.com/tongguang/unity-cli-loop.git?path=/Packages/src"
```

**从 tarball 安装**：在 [Releases](https://github.com/tongguang/unity-cli-loop/releases) 里找到标记为 `v<版本号>` 的插件发布，下载 `io.github.tongguang.tloop-<版本号>.tgz`，然后在 Package Manager 里选择 **Add package from tarball**。

**在 Editor 里安装 CLI 和 skills**：插件导入后，也可以用 `Window > Tuanjie CLI Loop > Settings` 里的 **Install CLI** 和 **Install Skills** 按钮安装。

**skills 的安装位置**：skills 默认装在团结工程根目录下。如果你在 Git 仓库根目录启动 AI 工具，而团结工程在子目录里，加上 `--root git`，skills 会装到仓库根目录。加上 `--global` 则装到用户目录，所有工程共用。

本项目只通过 GitHub 分发，没有发布到 OpenUPM、Homebrew 或 winget。

## 兼容性

已在 Windows 上验证：

- **团结 1.9.1**（Editor 版本 `2022.3.62t9`，也是本仓库开发工程使用的版本）：图形界面的 Editor 连接正常，编译和定向 EditMode 测试通过。
- **团结 1.10.3**（Editor 版本 `2022.3.62t15`）：在 batchmode 运行的 Editor 中，插件导入、连接、定向测试和离线编译检查通过。

macOS 和 Linux 上的团结 Editor、其他团结版本都还没有测试。`tloop launch`、PlayMode、热重载和录屏在团结上还需要进一步验证。测试明细和已知问题见 [团结兼容性说明](docs/tuanjie-compatibility.md)。

## 常见问题

**`tloop launch` 找不到 Editor**

`tloop launch` 读取 `ProjectSettings/ProjectVersion.txt` 里的 `m_EditorVersion`（形如 `2022.3.62t15`），再到团结 Hub 的安装记录和默认安装目录里找对应的 Editor。注意这个值和团结的产品版本号（形如 `1.10.3`）是两回事。

从 Unity 迁移过来的工程，先用团结打开并保存一次，让这个文件记录团结的版本。Editor 装在其他位置时，直接指定可执行文件：

```powershell
tloop launch --editor-path 'D:\Engines\Tuanjie\Editor\Tuanjie.exe'
```

**命令返回 Busy**

一个 Editor 同一时间只执行一条 tloop 命令。等前一条结束即可。如果它一直不结束，用 `tloop cancel` 取消。

**Windows 上 tloop 突然退出，或者报 `UNITY_NOT_REACHABLE`**

可能是 Microsoft Defender 误报，把 `tloop.exe` 隔离了。确认和恢复的方法见 [Windows 兼容性说明](docs/windows-compatibility.md#microsoft-defender-false-positives)。

**在沙箱里运行时连不上 Editor**

AI 工具在沙箱里执行命令时（例如开启了沙箱的 Claude Code），本地通信可能被拦截，表现为 Editor 运行正常但命令连不上。排查方法见 [Claude Code 沙箱与 IPC](docs/claude-code-sandbox.md)。

## 工作原理

`tloop` 和 Editor 在本机直接通信：Windows 上用命名管道，macOS / Linux 上用 Unix domain socket，不监听网络端口。每个工程有自己的连接，同时打开多个工程也不会互相干扰。

`tloop` 找到工程后，把命令交给与插件版本匹配的 project runner，由它和 Editor 通信。插件里记录了需要的 runner 版本，首次在工程里运行命令时，`tloop` 会从 GitHub Releases 下载这个版本并缓存。CLI 的发行包会经过校验和与来源证明（attestation）验证。

## 扩展与文档

- [自定义工具](Packages/src/Documentation~/custom-tools.md)：用 C# 编写自己的 tloop 命令，并附上 skill。程序集和扩展 API 沿用上游的名称，方便迁移已有的自定义工具。
- [团结兼容性说明](docs/tuanjie-compatibility.md)：验证明细、Editor 查找规则，以及与 `uloop` 的隔离方式。
- [Project runner pin](docs/project-runner-pin.md)：`tloop` 如何为工程选择 project runner 版本。
- [插件发布流程](docs/unity-package-release.md)
- [上游演示视频](https://github.com/hatayama/unity-cli-loop#demos)：由上游作者用原版 Unity 和 `uloop` 录制，不代表团结上的效果。

## 致谢与许可

`tloop` 的命令、skills 和整套开发循环都来自上游的 [Unity CLI Loop](https://github.com/hatayama/unity-cli-loop)，感谢 Masamichi Hatayama 的工作。本分支主要做团结适配：Editor 查找、与 `uloop` 的隔离，以及独立的发布流程。仓库里的 logo、截图和演示素材也来自上游。

本项目沿用上游的 MIT 许可证，版权声明为 Copyright (c) 2025 Masamichi Hatayama，全文见 [LICENSE.md](LICENSE.md)。插件内第三方组件的许可见 [THIRD-PARTY-NOTICES.md](Packages/src/THIRD-PARTY-NOTICES.md)。
