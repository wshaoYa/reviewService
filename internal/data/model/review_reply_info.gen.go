// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.

package model

import (
	"time"
)

const TableNameReviewReplyInfo = "review_reply_info"

// ReviewReplyInfo 评价商家回复表
type ReviewReplyInfo struct {
	ID        int64     `gorm:"column:id;primaryKey;autoIncrement:true;comment:主键" json:"id"`                      // 主键
	CreateBy  string    `gorm:"column:create_by;not null;comment:创建方标识" json:"create_by"`                          // 创建方标识
	UpdateBy  string    `gorm:"column:update_by;not null;comment:更新方标识" json:"update_by"`                          // 更新方标识
	CreateAt  time.Time `gorm:"column:create_at;not null;default:CURRENT_TIMESTAMP;comment:创建时间" json:"create_at"` // 创建时间
	UpdateAt  time.Time `gorm:"column:update_at;not null;default:CURRENT_TIMESTAMP;comment:更新时间" json:"update_at"` // 更新时间
	DeleteAt  time.Time `gorm:"column:delete_at;comment:逻辑删除标记" json:"delete_at"`                                  // 逻辑删除标记
	Version   int32     `gorm:"column:version;not null;comment:乐观锁标记" json:"version"`                              // 乐观锁标记
	ReplyID   int64     `gorm:"column:reply_id;not null;comment:回复id" json:"reply_id"`                             // 回复id
	ReviewID  int64     `gorm:"column:review_id;not null;comment:评价id" json:"review_id"`                           // 评价id
	StoreID   int64     `gorm:"column:store_id;not null;comment:店铺id" json:"store_id"`                             // 店铺id
	Content   string    `gorm:"column:content;not null;comment:评价内容" json:"content"`                               // 评价内容
	PicInfo   string    `gorm:"column:pic_info;not null;comment:媒体信息：图片" json:"pic_info"`                          // 媒体信息：图片
	VideoInfo string    `gorm:"column:video_info;not null;comment:媒体信息：视频" json:"video_info"`                      // 媒体信息：视频
	ExtJSON   string    `gorm:"column:ext_json;not null;comment:信息扩展" json:"ext_json"`                             // 信息扩展
	CtrlJSON  string    `gorm:"column:ctrl_json;not null;comment:控制扩展" json:"ctrl_json"`                           // 控制扩展
}

// TableName ReviewReplyInfo's table name
func (*ReviewReplyInfo) TableName() string {
	return TableNameReviewReplyInfo
}
