# Tuanjie CLI Loop (`tloop`)

[简体中文](README.md) | 日本語 | [Upstream (English)](https://github.com/hatayama/unity-cli-loop)

AI エージェントから CLI を通じて Tuanjie（团结引擎）のコンパイル、テスト、ログ確認、Editor 操作を行うためのプロジェクトです。

本プロジェクトは **Masamichi Hatayama** 氏の [Unity CLI Loop (`uloop`)](https://github.com/hatayama/unity-cli-loop) を基に、[tongguang](https://github.com/tongguang) が Tuanjie 専用の派生版として保守しています。上流の開発ループ、ツール実装、skills を基盤に、Editor の自動検出、実行環境の分離、配布フローを変更しています。

## 対応範囲

- 本ブランチは Tuanjie 専用です。通常の Unity プロジェクトには上流の `uloop` を使用してください。
- CLI は `tloop`、runner は `tloop-project-runner`、UPM パッケージは `io.github.tongguang.tloop` です。
- 環境変数 `TLOOP_*`、キャッシュ、`.tloop/`、`tloop-*` skills は上流と分離しています。同じマシンで共存できます。
- Windows は名前付きパイプ、macOS / Linux は Unix domain socket を使用します。TCP ポートの設定は不要です。
- `UnityEngine`、`UnityEditor`、既存のアセンブリ名と拡張 API 名は維持しています。

Windows で 2 つの Tuanjie バージョンを確認しました。

- **Tuanjie 1.10.3 (`2022.3.62t15`)**: ヘッドレス環境でのインポート、IPC 接続、90 件の限定テスト、57 アセンブリのオフラインコンパイル。
- **Tuanjie 1.9.1 (`2022.3.62t9`)**（このリポジトリの開発プロジェクトが使うバージョン）: Tuanjie Hub から開いた GUI の Editor で IPC 接続とエラー・警告ゼロのコンパイルを確認し、987 件の限定 EditMode テストのうち 985 件が成功、1 件がスキップでした。唯一の失敗は、この環境でクリップボードの読み戻しが空になったことによるものです。

`tloop launch` による起動、PlayMode、hot reload、録画などは追加検証が必要です。他の Tuanjie バージョンおよび macOS / Linux の Editor 対応は未検証です。Unity の CI 結果は Tuanjie の互換性を保証しません。[検証の詳細と既知の制限](docs/tuanjie-compatibility.md) を参照してください。

## リリース

| コンポーネント | バージョン |
| --- | --- |
| Tuanjie UPM パッケージ | [3.22.1](https://github.com/tongguang/unity-cli-loop/releases/tag/v3.22.1) |
| CLI dispatcher | [3.14.1](https://github.com/tongguang/unity-cli-loop/releases/tag/dispatcher-v3.14.1) |
| Project runner | [3.13.1](https://github.com/tongguang/unity-cli-loop/releases/tag/tloop-project-runner-v3.13.1) |

runner のバージョンはプロジェクトの pin から自動選択されます。

## インストール

### 1. CLI

Windows PowerShell:

```powershell
irm https://raw.githubusercontent.com/tongguang/unity-cli-loop/main/scripts/install.ps1 | iex
```

macOS / Linux / Windows Git Bash:

```sh
curl -fsSL https://raw.githubusercontent.com/tongguang/unity-cli-loop/main/scripts/install.sh | sh
```

`tloop --version` で確認できます。Homebrew、winget、OpenUPM での配布はまだ設定していません。`tloop package install` は OpenUPM への登録が必要なため、今回は次の Git URL または tarball を使用してください。

### 2. Tuanjie パッケージ

Tuanjie の `Window > Package Manager` で **Add package from git URL** を選択し、次の URL を入力してください。

```text
https://github.com/tongguang/unity-cli-loop.git?path=/Packages/src#v3.22.1
```

または [UPM リリース](https://github.com/tongguang/unity-cli-loop/releases/tag/v3.22.1) の `.tgz` をダウンロードし、**Add package from tarball** でインストールできます。通常の UPM パッケージとして配布しており、Unity Cloud のパッケージ署名は不要です。CLI のチェックサムと来歴証明の検証は維持しています。

### 3. Skills

パッケージには `tloop-*` の skill ソースも含まれています。パッケージのインストール後、Tuanjie プロジェクトのルートで実行してください。

```sh
# Claude Code
tloop skills install --claude

# Codex and other tools that read .agents/skills
tloop skills install --agents
```

Editor の `Window > Tuanjie CLI Loop > Settings` から **Install CLI**、**Install Skills** を使うこともできます。既存の `uloop-*` skills とは別名でインストールされます。

### 4. 接続と操作

最初に Tuanjie でプロジェクトを開いて保存し、`ProjectSettings/ProjectVersion.txt` に Tuanjie の Editor バージョンを記録してください。プロジェクトのルートで使用します。

```sh
tloop status
tloop list
tloop compile
tloop get-logs --max-count 10
tloop run-tests --filter-type class --filter-value YourTestClass
```

Editor が起動していなければ `tloop launch` を使用できます。プロジェクトに記録された `m_EditorVersion` と Tuanjie Hub の登録情報、既定のインストール先から Editor を探します。`2022.3.62t15` が Editor バージョン、`1.10.3` が製品バージョンです。

各コマンドの引数は `tloop <command> --help`、利用例と全体の説明は [中国語 README](README.md) を参照してください。Linux では `focus-window` は利用できません。

## 参考

- [Tuanjie 互換性と環境の分離](docs/tuanjie-compatibility.md)
- [カスタムツールの開発](Packages/src/Documentation~/custom-tools.md)
- [Project runner pin](docs/project-runner-pin.md)
- [UPM リリースのフロー](docs/unity-package-release.md)
- [上流のデモ](https://github.com/hatayama/unity-cli-loop#demos): 上流作者による Unity / `uloop` の映像です。Tuanjie の検証結果ではありません。リポジトリ内のロゴ、スクリーンショット、デモ素材も上流由来です。上流ロゴには Daft Punk の *Human After All* のジャケットに着想を得たとの注記があります。

## ライセンスと謝辞

上流の **MIT License** を継承しています。

> Copyright (c) 2025 Masamichi Hatayama

著作権表示、許諾条件、免責条項の全文は [LICENSE.md](LICENSE.md) に保持しています。名称の変更や Tuanjie 対応によって、上流コードの著作者の帰属は変わりません。コピー、変更、再配布の際には、元の著作権表示とライセンス全文を保持してください。

同梱の第三者コンポーネントについては [THIRD-PARTY-NOTICES.md](Packages/src/THIRD-PARTY-NOTICES.md) と各プラグインのライセンスも参照してください。
