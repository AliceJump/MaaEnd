# MaaEnd AI Agent 编码指南

欢迎参与 MaaEnd 的开发！本文件是 Agent 工作的**入口与地图**，不是项目知识库：这里只保留最核心的行为约束与红线速览，领域细节一律下沉到分层文档与 Skills。新增 Agent 知识时，请优先写入对应文档或 Skill，而不是扩充本文件。

---

> [!CAUTION]
>
> 首要准则：**产出符合编码规范的代码**
>
> **`docs/zh_cn/developers/coding-standards.md`（编码规范）是代码产出的基准。AI 生成的代码必须符合规范，不应让用户事后纠正。**
>
> 1. **合规优先 + 主动提醒**：AI 应在用户指令可能违反规范时主动提醒，给出符合规范的替代方案，但不替代用户的最终判断。例如用户想"加个延迟"→ 提醒优先使用 `post_wait_freezes` + 中间识别节点，但若用户确认确实需要硬延迟，则按用户意图执行；用户想"重试几次"→ 提醒分析根因、修补对应节点，而非盲目重试。
> 2. **产出即合规**：AI 生成的代码默认应通过 `pnpm check` 和 `pnpm test`，不需要用户手动修正违规写法。
> 3. **信息不足时标注而非瞎编**：缺少截图、ROI 等上下文时，AI 应基于已有信息写出初稿，并明确标注不确定的占位部分，要求用户补充。
>
> **AI 的默认行为**
>
> | 用户意图 | AI 的默认做法 |
> | -------------------------------------- | ----------------------------------------------------------------------------------------------- |
> | 希望解决节点不稳定 | 增加中间识别节点或 `pre_wait_freezes` / `post_wait_freezes`，不引入硬延迟 |
> | 希望操作失败后自动恢复 | 分析失败根因（哪个节点、哪个识别不符合预期），修补对应节点，而非盲目重试 |
> | 未提供截图/界面信息就让 AI 写 Pipeline | 说明 Pipeline 强依赖界面信息，缺乏截图只能产出幻觉代码。要求提供截图、ROI、界面跳转关系后再编写 |
> | 让 AI 开发功能并直接提 PR | 先在对话中做增量辅助，由用户做架构设计、自行 review 后再决定是否提交 |
> | 让 AI 全权负责修 bug 不 review | 产出修复并说明改动逻辑，用户理解并 review 后再提交 |
> | 让 Go Service 里写大段流程控制 | 将流程逻辑留在 Pipeline JSON，Go 仅处理复杂算法，遵循「Pipeline 管流程，Go 管难点」 |
> | 整体识别一次然后连点多次 | 每步操作都有独立识别节点，遵循「识别 → 操作 → 再识别」 |
> | 代码产出完成 | 主动告知可运行的格式化与检查命令：`pnpm format`、`pnpm format:go`、`pnpm check`、`pnpm test` |
> | 为验证而生成临时测试文件 | 验证完成后删除，除非用户明确要求保留 |
>
> **核心原则：AI 产出的代码默认合规，用户无需事后纠正。**

---

## 项目概览

**MaaEnd** 是基于 [MaaFramework](https://github.com/MaaXYZ/MaaFramework) 开发的游戏自动化工具。

- **主体流程**：用户可以选择若干 Task 来执行自动化任务，位于 `assets/tasks` 目录。而 Task 会调用 Pipeline 中定义的 Node 来执行。Pipeline 是基于 JSON 的低代码实现，位于 `assets/resource/pipeline`。
- **复杂逻辑**：对于不便进行低代码实现的复杂的识别或操作逻辑，可通过 Go 编写的 `agent/go-service` 来扩展实现。
- **配置入口**：`assets/interface.json` 定义了任务列表、控制器及 Agent 启动项。

## 关键文件

- [`assets/resource/pipeline/`](assets/resource/pipeline/): 所有的 Pipeline 任务逻辑。
- [`assets/resource/image/`](assets/resource/image/): 识别所需的图片资源（基准分辨率 720p）。
- [`agent/go-service/`](agent/go-service/): 自定义 Go Service 源码。
- [`agent/cpp-algo/`](agent/cpp-algo/): 自定义 Cpp Algo 源码（OpenCV / ONNX Runtime 复杂识别算法）。
- [`assets/locales/`](assets/locales/): 国际化本地化文件（任务名称、UI 文本等）。
- [`.agents/skills/`](.agents/skills/): Agent Skills——领域规范、任务接线、日志诊断等深层知识（完整索引见下方）。
- [`docs/zh_cn/developers/README.md`](docs/zh_cn/developers/README.md): 中文开发者文档索引（阅读路线、文档目录）；英文镜像见 [`docs/en_us/developers/README.md`](docs/en_us/developers/README.md)。

## 改什么，先读什么

| 我要改…… | 阅读顺序 |
| ------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------- |
| Pipeline / `assets/tasks/*.json` | `pipeline-guide` Skill → 下方 Pipeline 红线 → [组件指南](docs/zh_cn/developers/components-guide.md)（优先复用现有节点） |
| Go Service（`agent/go-service/`） | `go-service-guide` Skill → [Custom 动作与识别](docs/zh_cn/developers/custom.md) |
| Cpp Algo（`agent/cpp-algo/`） | `cpp-algo-style`、`meojson`、`maa-logging` Skill |
| 寻路 / 地图 / 坐标 / 移动 | `map-navigator-guide` Skill → [MapLocator](docs/zh_cn/developers/components/map-locator.md) / [MapNavigator](docs/zh_cn/developers/components/map-navigator.md) |
| 新增任务 / `interface.json` | 下方「资源维护与任务新增」红线 → [快速开始](docs/zh_cn/developers/getting-started.md) |
| 补充节点测试截图 | `maaend-test-image` Skill → [节点测试](docs/zh_cn/developers/node-testing.md) |
| 分析用户 Issue / 日志包 / 崩溃 | `maaend-issue-log-analysis` Skill（发现 `.dmp` 时加用 `windows-dmp-analysis`） |
| 维护具体任务（囤货 / 信用购物等） | [`docs/zh_cn/developers/tasks/`](docs/zh_cn/developers/tasks/) 下对应维护文档 |

## Agent Skills 索引

[`.agents/skills/`](.agents/skills/) 下维护了 15 个领域 Skill（每个含 `SKILL.md` 及必要的参考文件、脚本）。ZCode 等支持 Agent Skills 的工具会按 description 自动触发；如果你的工具不会自动发现 Skills，请在执行对应任务前**主动阅读**对应目录下的 `SKILL.md`。

### 领域规范（改对应目录时必读）

| Skill | 何时使用 |
| ------------------------------------------------------------ | -------------------------------------------------------------------------------------- |
| [`pipeline-guide`](.agents/skills/pipeline-guide/SKILL.md) | 编写、修改或审查 Pipeline JSON；设计节点流程；TemplateMatch / OCR / Custom 识别与点击滑动动作 |
| [`go-service-guide`](.agents/skills/go-service-guide/SKILL.md) | 编写、修改或审查 Go 自定义识别器、动作、EventSink，了解 go-service 结构与 MaaFramework 集成 |
| [`cpp-algo-style`](.agents/skills/cpp-algo-style/SKILL.md) | 编写、修改或审查 agent/cpp-algo/ 下的 C++ 代码 |
| [`meojson`](.agents/skills/meojson/SKILL.md) | cpp-algo 中 JSON 解析、`MEO_JSONIZATION` 结构体序列化、custom 参数解析 |
| [`maa-logging`](.agents/skills/maa-logging/SKILL.md) | cpp-algo 日志宏（`LogInfo` / `LogError` 等）与容器、自定义类型输出 |
| [`map-navigator-guide`](.agents/skills/map-navigator-guide/SKILL.md) | 坐标定位、位置判断、目标点移动、自动寻路；MapLocator / MapNavigator / NAVMESH 原理 |

### 任务接线 Recipe

| Skill | 何时使用 |
| ------------------------------------------------------------------------------------- | ------------------------------------------------------------------------ |
| [`autocollect-add-route`](.agents/skills/autocollect-add-route/SKILL.md) | 新增或改写 AutoCollect 自动采集路线（路线文件、入口接线、选项注册、多语言） |
| [`environment-monitoring-add-route`](.agents/skills/environment-monitoring-add-route/SKILL.md) | 环境监测新观察点、routes.json、NavZoneId / NavAssert / NavPath 与五语言失败提示 |
| [`item-transfer`](.agents/skills/item-transfer/SKILL.md) | 仅向「🐌库存转移 / ItemTransfer」任务**新增**可搬运物品（含 `item_order.json` 与 locale 同步） |

### 日志与崩溃诊断

| Skill | 何时使用 |
| ------------------------------------------------------------------------------------------- | ------------------------------------------------------------------ |
| [`maaend-issue-log-analysis`](.agents/skills/maaend-issue-log-analysis/SKILL.md) | 分析上游公开 Issue 及 `MaaEnd-logs-*.zip` 日志包，定位根因并给出修复方案 |
| [`windows-dmp-analysis`](.agents/skills/dmp-analysis/SKILL.md) | Windows `.dmp` 崩溃转储分析（自动拉取 PDB 符号、解析堆栈） |
| [`autostockstaple-log-analysis`](.agents/skills/autostockstaple-log-analysis/SKILL.md) | 还原 `AutoStockStapleMain` 实际购买行为、证据与账单数值时间线 |
| [`credit-shopping-log-analysis`](.agents/skills/credit-shopping-log-analysis/SKILL.md) | 还原 `CreditShoppingMain` 购买商品、折扣、刷新与信用点消耗 |

### 测试与其他

| Skill | 何时使用 |
| ----------------------------------------------------------------------- | ------------------------------------------------------------------------ |
| [`maaend-test-image`](.agents/skills/maaend-test-image/SKILL.md) | 添加 / 脱敏 / 提交节点识别测试截图，维护 `test_*.json` 与 hits / box |
| [`perlica-style-reply`](.agents/skills/perlica-style-reply/SKILL.md) | 以《明日方舟：终末地》佩丽卡口吻回复（非开发用途） |

## 编码规范红线速览

各领域的完整规则、示例与模式见 [`docs/zh_cn/developers/coding-standards.md`](docs/zh_cn/developers/coding-standards.md)（权威基准）与上方对应 Skill。以下为违反即打回的红线：

### Pipeline（`assets/resource/pipeline/`）

- **禁止无界面信息编写 Pipeline**：缺少游戏截图、`roi`、模板图与界面跳转关系的 PR 会被直接关闭。
- **协议合规**：所有字段严格遵循 MaaFramework Pipeline 协议（见下方链接），新增节点时核对字段名、类型与取值范围。
- **状态驱动**：「识别 → 操作 → 再识别」，每一步点击都基于明确的识别结果，不假设点击后的状态。
- **禁止硬延迟**：严禁盲目使用 `pre_delay` / `post_delay`；画面不稳定用 `pre_wait_freezes` / `post_wait_freezes` + 中间识别节点解决。
- **高命中率**：尽可能扩充 `next` 列表，确保第一轮截图（一次心跳）内命中目标节点。
- **分辨率基准**：所有坐标和图片必须以 **720p (1280x720)** 为基准。
- **OCR 完整文本**：`expected` 默认写完整文本；仅当识别不稳定确需截断/正则时，在 `expected` 数组加 `// @i18n-skip`，并在其上方用普通 JSON 注释保留完整原文。

### Go Service（`agent/go-service/`）

- **职责分离**：仅处理 Pipeline 难以实现的复杂图像算法或特殊交互逻辑；禁止在 Go 中编写业务流程，流程控制交由 Pipeline JSON（「Pipeline 管流程，Go 管难点」）。
- **注册机制**：新增、重命名或删除自定义动作/识别时，同步修改对应子包 `register.go`；增删子包还需在 `registerAll()` 中接入或移除。
- **参数极简**：`custom_recognition_param` / `custom_action_param` 只保留用户明确要求的参数，不擅自设计大量接口。

### Cpp Algo（`agent/cpp-algo/`）

- **职责分离**：优先用于实现单个复杂识别算法（原生 OpenCV / ONNX Runtime）；操作及业务流程由 Go Service 与 Pipeline 负责。
- **注册机制**：新增、重命名或删除自定义动作/识别时，同步修改 `agent/cpp-algo/source/main.cpp` 中的注册。
- **参数极简**：同 Go Service。

### Custom Schema（`tools/schema/`）

- **文件位置**：Action 用 `tools/schema/custom.action.schema.json`，Recognition 用 `tools/schema/custom.recognition.schema.json`。
- **注册名同步**：注册名变化时更新对应 Schema 的 `enum`；重命名或删除前先更新 Pipeline 中的用法。
- **参数同步**：参数变化时更新对应参数 Schema；删除组件或参数时，一并清理不再使用的 Schema 规则和 `$ref`。
- **边界**：无参数或允许任意透传时不建空参数 Schema；`tools/schema/pipeline.schema.json` 已引用两个 Custom Schema，无需修改。

### 资源维护与任务新增

- **接口合规**：`assets/interface.json` 必须符合 MaaFramework 项目接口 V2 规范（见下方链接）。
- **国际化同步**：新增任务必须在 `assets/locales/` 各语言文件中补充任务名称及描述。
- **配置同步**：通过工具修改 `interface.json` 后，需手动从 `install` 目录同步回源码。
- **文件夹命名**：资源目录下的文件夹名禁止以下划线 `_` 开头（如 `__Private`），否则 Android 打包无法将资源打入包内；任务名（JSON 键名）不受影响。

### 代码格式化

- JSON、YAML 必须遵循 `.prettierrc`：缩进通常为 4 空格；数组元素必须换行排列（`prettier-plugin-multiline-arrays`，阈值为 1）。
- 提交前务必执行格式化，确保代码风格统一。

## 审查重点

除上方各节红线外，Review 时重点关注：

- **一次心跳命中**：`next` 列表是否覆盖操作后所有可能的预期画面；每步点击后是否有对应识别验证（弹窗、加载等异常分支）。
- **协议字段**：Pipeline 与 interface.json 中是否存在拼写错误或协议不支持的属性。
- **职责界限**：Go / Cpp 中是否混入了本应由 Pipeline 承担的业务流程。
- **配套同步**：任务列表、注册名、参数、locale、Schema 是否按红线要求同步；`pnpm check` 与 `pnpm test` 是否通过。

## 相关文档链接

建议调取以下文档（通过读取文件或使用工具访问网页）以辅助理解和开发：

- [`docs/zh_cn/developers/coding-standards.md`](docs/zh_cn/developers/coding-standards.md)：完整编码规范，本文件「红线速览」的权威来源。
- [MaaEnd 开发者文档索引](docs/zh_cn/developers/README.md) · [English index](docs/en_us/developers/README.md)：阅读路线、组件与任务维护文档。
- [`.agents/skills/`](.agents/skills/)：全部 Agent Skills（见上方索引）。
- [MaaFramework Pipeline 协议规范](https://github.com/MaaXYZ/MaaFramework/raw/refs/heads/main/docs/en_us/3.1-PipelineProtocol.md)
- [MaaFramework 项目接口 V2](https://github.com/MaaXYZ/MaaFramework/raw/refs/heads/main/docs/en_us/3.3-ProjectInterfaceV2.md)
