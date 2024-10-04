package biz

import (
	"reviewService/internal/data/model"
	"strings"
	"time"
)

type MyTime time.Time

// MyReviewInfo 为解决时间格式不兼容 自行嵌套一下
type MyReviewInfo struct {
	*model.ReviewInfo
	CreateAt     MyTime `json:"create_at"`
	UpdateAt     MyTime `json:"update_at"`
	Anonymous    int32  `json:"anonymous,string"` // es中数据均为string类型，序列化反序列化时以string为准
	Score        int32  `json:"score,string"`
	ServiceScore int32  `json:"service_score,string"`
	ExpressScore int32  `json:"express_score,string"`
	HasMedia     int32  `json:"has_media,string"`
	Status       int32  `json:"status,string"`
	IsDefault    int32  `json:"is_default,string"`
	HasReply     int32  `json:"has_reply,string"`
	ID           int64  `json:"id,string"`
	Version      int32  `json:"version,string"`
	ReviewID     int64  `json:"review_id,string"`
	OrderID      int64  `json:"order_id,string"`
	SkuID        int64  `json:"sku_id,string"`
	SpuID        int64  `json:"spu_id,string"`
	StoreID      int64  `json:"store_id,string"`
	UserID       int64  `json:"user_id,string"`
}

// UnmarshalJSON 实现反序列化时的接口
func (t *MyTime) UnmarshalJSON(b []byte) error {
	s := string(b)
	trim := strings.Trim(s, `"`)

	tmp, err := time.Parse(time.DateTime, trim)
	if err != nil {
		return err
	}

	*t = MyTime(tmp)
	return nil
}
