package utils

import (
	"DiTing-Go/global"
	"strconv"
	"time"
)

func TimestampStrToTimeStr(str *string) (*string, error) {
	if str != nil && *str != "" {
		// 时间戳转时间
		timestamp, err := strconv.ParseInt(*str, 10, 64)
		if err != nil {
			global.Logger.Errorf("时间戳转换失败 %s", err)
			return nil, err
		}
		cursor := time.Unix(0, timestamp)
		cursorStr := cursor.Format(time.RFC3339Nano)
		return &cursorStr, nil
	}
	return nil, nil
}
