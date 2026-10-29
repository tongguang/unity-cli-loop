# Tuanjie CLI Loop (`tloop`)

[简体中文](README.md) | 日本語 | [Upstream (English)](https://github.com/hatayama/unity-cli-loop)

[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE.md)

`tloop` は、Claude Code や Codex などの AI コーディングアシスタントから Tuanjie（团结引擎）の Editor を直接操作するための CLI です。コンパイル、テストの実行、Console ログの確認、スクリーンショット、C# の実行ができます。

AI はコードを変更したあと自分でコンパイルし、エラーが出ればそのまま修正を続けます。コンパイルが通り、テストが成功するまでこれを繰り返します。Editor のエラーをチャットに貼り付ける作業は要りません。

本プロジェクトは Masamichi Hatayama 氏の [Unity CLI Loop](https://github.com/hatayama/unity-cli-loop)（`uloop`）を Tuanjie 向けにしたもので、[tongguang](https://github.com/tongguang) が保守しています。通常の Unity プロジェクトには上流の `uloop` を使ってください。両方を同じマシンに入れても干渉しません。

## できること

- **コンパイルとテスト**：プロジェクトをコンパイルしてエラーと警告を返します。EditMode / PlayMode テストの実行、Console の取得と消去もできます。
- **変更をすばやく確かめる**：hot reload は C# の変更をドメインリロードなしで実行中の Editor に反映し、PlayMode の状態も保ちます。オフラインコンパイルチェックは Editor に接続せずにコンパイルエラーを見つけます。
- **シーンを見る**：シーン階層の出力、GameObject の検索、Editor ウィンドウや Game View のスクリーンショット、録画ができます。
- **実行時の挙動を調べる**：PlayMode の操作、キーボードとマウス入力のシミュレーション、指定したソース行での一時停止とローカル変数の読み取り、その行の実行回数の計測ができます。
- **任意の C# を実行**：既存のコマンドでできない操作は、AI に Editor 内で C# を直接実行させられます。
- **拡張**：C# で独自のコマンドを作り、AI 向けの skill を添えられます。

各コマンドには skill が付いていて、いつ、どう使うかを AI に教えます。skills を入れておけば、やりたいことを普段の言葉で伝えるだけで、AI が使うコマンドを選びます。

## クイックスタート

必要なもの：Tuanjie Editor と、skills を読める AI ツール（Claude Code、Codex など）。

**1. CLI をインストール**

Windows（PowerShell）:

```powershell
irm https://raw.githubusercontent.com/tongguang/unity-cli-loop/main/scripts/install.ps1 | iex
```

macOS / Linux / Windows Git Bash:

```sh
curl -fsSL https://raw.githubusercontent.com/tongguang/unity-cli-loop/main/scripts/install.sh | sh
```

`tloop --version` で確認します。コマンドが見つからない場合は、ターミナルを開き直してください。

**2. プロジェクトにパッケージを追加**

Tuanjie プロジェクトのルートで実行します。

```sh
tloop package install
```

`Packages/manifest.json` にパッケージが追加されます。続いて Tuanjie でプロジェクトを開き（開いている場合は Editor に切り替え）、インポートが終わるのを待ちます。

**3. skills をインストール**

skills はパッケージに含まれているため、インポートが終わってからインストールします。同じくプロジェクトのルートで実行します。

```sh
# Claude Code
tloop skills install --claude

# Codex など .agents/skills を読むツール
tloop skills install --agents
```

**4. 試す**

```sh
tloop status
```

出力に `"State": "Ready"` があれば、CLI は Editor に接続できています。AI ツールにこう頼んでみてください。

> tloop でこのプロジェクトをコンパイルして、エラーがあれば直して。

## よく使うコマンド

普段は AI が呼び出しますが、ターミナルから手動で実行することもできます。

```sh
tloop compile
tloop get-logs --max-count 10
tloop run-tests --filter-type class --filter-value YourTestClass
tloop screenshot
tloop execute-dynamic-code --code 'using UnityEngine; Debug.Log("Hello from tloop!");'
```

| 用途 | コマンド |
| --- | --- |
| Editor の状態確認、実行中コマンドの取り消し | `status`、`cancel` |
| Editor の起動・再起動 | `launch` |
| コンパイル、オフラインコンパイルチェック | `compile`、`compile-check` |
| テストと Console | `run-tests`、`get-logs`、`clear-console` |
| hot reload | `hot-reload` |
| シーンとオブジェクト | `get-hierarchy`、`find-game-objects` |
| 任意の C# の実行 | `execute-dynamic-code` |
| ウィンドウ、スクリーンショット、録画 | `focus-window`、`screenshot`、`record-video` |
| Game View と Device Simulator | `set-game-view-size`、`game-view-mode` |
| PlayMode と入力 | `control-play-mode`、`simulate-keyboard`、`simulate-mouse-input`、`simulate-mouse-ui`、`replay-input` |
| ソース上の一時停止と変数のウォッチ | `enable-pause-point`、`await-pause-point`、`enable-watch`、`get-watch-values` |

`tloop list` で現在のプロジェクトで使える全コマンドを、`tloop <command> --help` で引数を確認できます。入力シミュレーションには Input System パッケージが必要です。Linux では `focus-window` を使えません。

## その他のインストール方法

**パッケージのバージョンを固定する**：`tloop package install` は既定で `main` ブランチを追跡します。リリース版に固定するには `tloop package install --version <バージョン>` を実行します。バージョンは [Releases](https://github.com/tongguang/unity-cli-loop/releases) で確認できます。

**Package Manager から追加する**：`Window > Package Manager` で **Add package from git URL** を選び、次の URL を入力します。

```text
https://github.com/tongguang/unity-cli-loop.git?path=/Packages/src
```

`Packages/manifest.json` の `dependencies` に次の項目を手で追加しても同じです。

```json
"io.github.tongguang.tloop": "https://github.com/tongguang/unity-cli-loop.git?path=/Packages/src"
```

**tarball からインストールする**：[Releases](https://github.com/tongguang/unity-cli-loop/releases) で `v<バージョン>` のパッケージリリースを開き、`io.github.tongguang.tloop-<バージョン>.tgz` をダウンロードして、Package Manager の **Add package from tarball** を選びます。

**Editor から CLI と skills を入れる**：パッケージのインポート後は、`Window > Tuanjie CLI Loop > Settings` の **Install CLI** と **Install Skills** ボタンからもインストールできます。

**skills のインストール先**：skills は既定で Tuanjie プロジェクトのルートに入ります。AI ツールを Git リポジトリのルートで起動していて、Tuanjie プロジェクトがそのサブディレクトリにある場合は `--root git` を付けると、リポジトリのルートに入ります。`--global` を付けるとユーザーディレクトリに入り、すべてのプロジェクトで共有されます。

配布は GitHub のみで、OpenUPM、Homebrew、winget には公開していません。

## 互換性

Windows で確認済みです。

- **Tuanjie 1.9.1**（Editor バージョン `2022.3.62t9`、このリポジトリの開発プロジェクトが使うバージョン）：GUI の Editor で接続でき、コンパイルと対象を絞った EditMode テストが通りました。
- **Tuanjie 1.10.3**（Editor バージョン `2022.3.62t15`）：batchmode で動かした Editor で、パッケージのインポート、接続、対象を絞ったテスト、オフラインコンパイルチェックが通りました。

macOS / Linux の Tuanjie Editor と、ほかの Tuanjie バージョンは未検証です。`tloop launch`、PlayMode、hot reload、録画は Tuanjie での追加検証が必要です。詳細と既知の問題は [Tuanjie 互換性メモ](docs/tuanjie-compatibility.md) を参照してください。

## よくある質問

**`tloop launch` が Editor を見つけられない**

`tloop launch` は `ProjectSettings/ProjectVersion.txt` の `m_EditorVersion`（`2022.3.62t15` のような値）を読み、Tuanjie Hub のインストール情報と既定のインストール先から対応する Editor を探します。この値は Tuanjie の製品バージョン（`1.10.3` のような値）とは別物です。

Unity から移行したプロジェクトは、一度 Tuanjie で開いて保存し、このファイルに Tuanjie のバージョンを記録させてください。Editor を別の場所に入れている場合は、実行ファイルを直接指定します。

```powershell
tloop launch --editor-path 'D:\Engines\Tuanjie\Editor\Tuanjie.exe'
```

**コマンドが Busy を返す**

1 つの Editor が同時に実行できる tloop コマンドは 1 つだけです。前のコマンドが終わるのを待ってください。いつまでも終わらない場合は `tloop cancel` で取り消します。

**Windows で tloop が突然終了する、または `UNITY_NOT_REACHABLE` になる**

Microsoft Defender が `tloop.exe` を誤検知して隔離した可能性があります。確認と復元の手順は [Windows 互換性メモ](docs/windows-compatibility.md#microsoft-defender-false-positives) にあります。

**サンドボックス内で Editor に接続できない**

AI ツールがサンドボックス内でコマンドを実行している場合（サンドボックスを有効にした Claude Code など）、ローカル通信が拒否されることがあります。Editor は正常に動いているのにコマンドだけがつながらないときは、[Claude Code のサンドボックスと IPC](docs/claude-code-sandbox.md) を参照してください。

## 仕組み

`tloop` と Editor はローカルで直接通信します。Windows では名前付きパイプ、macOS / Linux では Unix domain socket を使い、ネットワークポートは開きません。接続はプロジェクトごとに分かれているので、複数のプロジェクトを同時に開いても干渉しません。

`tloop` はプロジェクトを見つけると、パッケージのバージョンに合った project runner にコマンドを渡し、runner が Editor と通信します。必要な runner のバージョンはパッケージに記録されていて、プロジェクトで初めてコマンドを実行したときに `tloop` が GitHub Releases からダウンロードしてキャッシュします。CLI のリリースはチェックサムと来歴証明（attestation）で検証されます。

## 拡張とドキュメント

- [カスタムツール](Packages/src/Documentation~/custom-tools.md)：C# で独自の tloop コマンドを作り、skill を添える方法です。アセンブリと拡張 API の名前は上流と同じなので、既存のカスタムツールを移行しやすくなっています。
- [Tuanjie 互換性メモ](docs/tuanjie-compatibility.md)：検証の詳細、Editor の探索ルール、`uloop` との分離方法。
- [Project runner pin](docs/project-runner-pin.md)：`tloop` がプロジェクトに合った project runner のバージョンを選ぶ仕組み。
- [パッケージのリリース手順](docs/unity-package-release.md)
- [上流のデモ動画](https://github.com/hatayama/unity-cli-loop#demos)：上流の作者が通常の Unity と `uloop` で撮影したもので、Tuanjie での動作を示すものではありません。

## 謝辞とライセンス

`tloop` のコマンド、skills、開発ループの仕組みはすべて上流の [Unity CLI Loop](https://github.com/hatayama/unity-cli-loop) に由来します。Masamichi Hatayama 氏に感謝します。このブランチで手を入れているのは主に Tuanjie 対応の部分で、Editor の探索、`uloop` との分離、独立したリリースフローです。リポジトリ内のロゴ、スクリーンショット、デモ素材も上流のものです。上流のロゴは Daft Punk のアルバム *Human After All* のジャケットに着想を得ています。

上流と同じ MIT ライセンスを継承しています。著作権表示は Copyright (c) 2025 Masamichi Hatayama で、全文は [LICENSE.md](LICENSE.md) にあります。パッケージに含まれるサードパーティ製コンポーネントのライセンスは [THIRD-PARTY-NOTICES.md](Packages/src/THIRD-PARTY-NOTICES.md) を参照してください。
