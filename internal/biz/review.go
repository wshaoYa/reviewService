package biz

import (
	"context"
	"errors"
	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
	v1 "reviewService/api/review/v1"
	"reviewService/internal/data/model"
	"reviewService/pkg/snowflake"
)

// ReviewRepo 评价repo
type ReviewRepo interface {
	Save(context.Context, *model.ReviewInfo) (*model.ReviewInfo, error) // C端 发布评价
	GetByOrderID(context.Context, int64) (*model.ReviewInfo, error)
	ListReviewByStoreID(ctx context.Context, storeID int64, offset int64, limit int64) ([]*MyReviewInfo, error) // C端 依商家ID获取评价列表

	AuditReview(context.Context, *model.ReviewInfo) error       // O端 审核评价
	AuditAppeal(context.Context, *model.ReviewAppealInfo) error // O端 审核申诉

	CreateReply(context.Context, *model.ReviewReplyInfo) error                              // B端 回复评价
	CreateAppeal(context.Context, *model.ReviewAppealInfo) (*model.ReviewAppealInfo, error) // B端 申诉评价
}

// ReviewUsecase 评价usecase
type ReviewUsecase struct {
	repo ReviewRepo
	log  *log.Helper
}

// NewReviewUsecase 评价usecase构造函数
func NewReviewUsecase(repo ReviewRepo, logger log.Logger) *ReviewUsecase {
	return &ReviewUsecase{repo: repo, log: log.NewHelper(logger)}
}

// CreateReview C端 创建评价
func (uc *ReviewUsecase) CreateReview(ctx context.Context, r *model.ReviewInfo) (*model.ReviewInfo, error) {
	//业务逻辑校验——判断此订单是否已评价过
	reviews, err := uc.repo.GetByOrderID(ctx, r.OrderID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		uc.log.Errorf("[biz] CreateReview GetByOrderID failed,err:%v \n", err)
		return nil, v1.ErrorInternalError("系统内部错误")
	}
	if reviews != nil {
		return nil, v1.ErrorOrderReviewed("订单%v 已被评价过", r.OrderID)
	}

	//log
	uc.log.WithContext(ctx).Infof("CreateReview - orderID: %v", r.OrderID)

	//生成评价ID
	r.ReviewID = snowflake.GenID()

	////默认未审核 已在sql表结构层面默认实现
	//r.Status = 10

	return uc.repo.Save(ctx, r)
}

// ListReviewByStoreID C端 依商家ID获取评价列表
func (uc *ReviewUsecase) ListReviewByStoreID(ctx context.Context, storeID int64, page int64, size int64) ([]*MyReviewInfo, error) {
	//参数校验
	page = max(page, 1)
	if size <= 0 || size >= 50 {
		size = 10
	}

	offset := (page - 1) * size
	limit := size
	return uc.repo.ListReviewByStoreID(ctx, storeID, offset, limit)
}

// CreateReply B端 回复评价
func (uc *ReviewUsecase) CreateReply(ctx context.Context, r *model.ReviewReplyInfo) error {
	r.ReplyID = snowflake.GenID()
	return uc.repo.CreateReply(ctx, r)
}

// AppealReview B端 申诉评价
func (uc *ReviewUsecase) AppealReview(ctx context.Context, r *model.ReviewAppealInfo) (*model.ReviewAppealInfo, error) {
	r.AppealID = snowflake.GenID()
	return uc.repo.CreateAppeal(ctx, r)
}

// AuditReview O端 审核评价
func (uc *ReviewUsecase) AuditReview(ctx context.Context, r *model.ReviewInfo) error {
	return uc.repo.AuditReview(ctx, r)
}

// AuditAppeal O端 审核申诉
func (uc *ReviewUsecase) AuditAppeal(ctx context.Context, r *model.ReviewAppealInfo) error {
	return uc.repo.AuditAppeal(ctx, r)
}
