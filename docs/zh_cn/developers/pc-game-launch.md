# PC 游戏启动

`GameLaunch` 是连接 Win32 控制器前执行的预任务，目前仅向国服官服资源开放。
在 MXU 的任务列表中添加并启用「启动 PC 游戏（国服官服）」，配置：

- 游戏程序：`Endfield.exe` 的绝对路径，不加引号。
- 官方启动器：安装根目录 `Launcher.exe` 的绝对路径，也接受版本目录下的 `Games.exe`。
  已有通过核验的 `Games.exe` 时可留空；启动器未运行时必须填写。
- 游戏启动参数：国服官服默认 `-launcher_lang=zh-cn -launcher_sub_channel=1`。
  填写值会完整替换默认参数，显式留空表示不带参数。

同时启用「游戏设置」时，将其排在「启动 PC 游戏」前面。
MXU 的普通前置程序无需再直接启动 `Endfield.exe`，避免重复拉起。
该预任务不会自动添加到既有任务列表。

程序先检查是否已有 `Endfield.exe`，有则跳过全部启动步骤。
否则查找正在运行的 `Games.exe`；没有时先打开已核验的官方启动器，按进程状态等待就绪。
使用同一个进程句柄查询实际映像路径、验证有效 Authenticode 签名和允许的签名者证书 SHA-1 指纹，
再通过 `STARTUPINFOEXW` / `PROC_THREAD_ATTRIBUTE_PARENT_PROCESS` 创建游戏。
PID 动态获取，工作目录是游戏目录，不修改游戏文件。

证书允许列表对应国服 1.6.3 的 Launcher 规则。证书轮换后应核实并更新允许列表；
不能降级到只校验文件名。国际服及渠道服需要独立核验启动器、证书规则和语言/渠道参数，
不能直接复用国服官服的默认值。

创建进程后，预任务最多等待 120 秒，确认该 PID 仍在运行且出现可见的 `UnityWndClass` 窗口。
这表示控制器可继续连接，不代表已验证登录或进入游戏场景。
失败时在 `debug/go-service.log` 记录 Windows 错误或进程退出码；
窗口等待失败还会从 `OneChannelLC.log`、`u8sdk_pc.log` 本次新增的尾部中，
采集 `Parent process` 来源检查和 `OnU8V2InitCallback` 初始化结果，不转储其他 SDK 数据。
日志不可读时会明确记录不可用，不把旧记录作为本次 SDK 结果。
失败不会关闭游戏或官方启动器，保留现场供用户检查。

命令行入口（最后一个参数为 PI 预任务选项 JSON）：

```text
agent/go-service --pretask GameLaunch {"GameLaunch":{"GamePath":"D:\\Games\\Endfield Game\\Endfield.exe","LauncherPath":"D:\\Launcher\\Launcher.exe","Arguments":"-launcher_lang=zh-cn -launcher_sub_channel=1"}}
```

调用方应将 JSON 整体作为一个参数传递，按所在 shell 的规则处理引号。
