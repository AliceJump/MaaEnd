package gamelaunch

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"
)

type sdkLog struct {
	path   string
	offset int64
}

type sdkLogs []sdkLog

func captureSDKLogs() sdkLogs {
	var logs sdkLogs
	for _, name := range []string{"OneChannelLC.log", "u8sdk_pc.log"} {
		path := filepath.Join(os.Getenv("USERPROFILE"), "AppData", "LocalLow", "Hypergryph", "Endfield", "sdklogs", name)
		var offset int64
		if info, err := os.Stat(path); err == nil {
			offset = info.Size()
		}
		logs = append(logs, sdkLog{path, offset})
	}
	return logs
}

// report 只记录本次启动追加的尾部，避免把上一次成功启动误报为当前 SDK 检查结果。
func (logs sdkLogs) report() {
	for _, entry := range logs {
		func() {
			file, err := os.Open(entry.path)
			if err != nil {
				log.Warn().Err(err).Str("sdk_log", entry.path).Msg("SDK diagnostics unavailable")
				return
			}
			defer file.Close()
			info, err := file.Stat()
			if err != nil {
				log.Warn().Err(err).Str("sdk_log", entry.path).Msg("cannot inspect SDK log")
				return
			}
			offset := entry.offset
			if info.Size() < offset {
				offset = 0 // SDK 在本次启动时截断了日志。
			}
			if info.Size()-offset > 8192 {
				offset = info.Size() - 8192
			}
			data, err := io.ReadAll(io.NewSectionReader(file, offset, info.Size()-offset))
			// SDK 还会输出账户等数据，只采集来源检查和初始化结果。
			var records []string
			for _, line := range strings.Split(string(data), "\n") {
				if strings.Contains(line, "Parent process") || strings.Contains(line, "OnU8V2InitCallback") {
					records = append(records, strings.TrimSuffix(line, "\r"))
				}
			}
			log.Error().Err(err).Str("sdk_log", entry.path).Strs("new_check_records", records).Msg("game launch SDK diagnostics")
		}()
	}
}
