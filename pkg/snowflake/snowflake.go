package snowflake

import (
	"errors"
	"time"
)
import sf "github.com/bwmarrin/snowflake"

var (
	InvalidInitParamErr  = errors.New("snowflake初始化失败，无效的startTime或machineID")
	InvalidTimeFormatErr = errors.New("snowflake初始化失败，无效的startTime格式")
)

var node *sf.Node

// Init 初始化雪花算法node
func Init(startTime string, machineID int64) (err error) {
	if len(startTime) == 0 || machineID < 0 {
		return InvalidInitParamErr
	}

	var st time.Time
	if st, err = time.Parse("2006-01-02", startTime); err != nil {
		return InvalidTimeFormatErr
	}

	//设置sf起始时间
	sf.Epoch = st.UnixNano() / 1000_000

	node, err = sf.NewNode(machineID)
	if err != nil {
		panic(err)
	}
	return
}

// GenID 生成分布式唯一ID
func GenID() int64 {
	return node.Generate().Int64()
}
