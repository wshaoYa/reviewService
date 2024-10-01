package data

import (
	"context"
	"errors"
	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
	v1 "reviewService/api/review/v1"
	"reviewService/internal/biz"
	"reviewService/internal/data/model"
	"reviewService/internal/data/query"
)

type reviewRepo struct {
	data *Data
	log  *log.Helper
}

// NewReviewRepo .
func NewReviewRepo(data *Data, logger log.Logger) biz.ReviewRepo {
	return &reviewRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// Save 创建评价
func (r *reviewRepo) Save(ctx context.Context, review *model.ReviewInfo) (*model.ReviewInfo, error) {
	return review, r.data.q.ReviewInfo.WithContext(ctx).Create(review)
}

// GetByOrderID 根据order id 获取评价
func (r *reviewRepo) GetByOrderID(ctx context.Context, orderID int64) (*model.ReviewInfo, error) {
	return r.data.q.WithContext(ctx).ReviewInfo.
		Where(r.data.q.ReviewInfo.OrderID.Eq(orderID)).
		First()
}

// AuditReview O端 审核评价
func (r *reviewRepo) AuditReview(ctx context.Context, review *model.ReviewInfo) error {
	ri := r.data.q.ReviewInfo
	_, err := ri.WithContext(ctx).
		Where(ri.ReviewID.Eq(review.ReviewID)).
		Updates(model.ReviewInfo{
			Status:    review.Status,
			OpReason:  review.OpReason,
			OpRemarks: review.OpRemarks,
			OpUser:    review.OpUser,
		})
	return err
}

// CreateReply  B端 回复评价
func (r *reviewRepo) CreateReply(ctx context.Context, reviewReply *model.ReviewReplyInfo) error {
	//业务校验--必须未回复过
	review, err := r.data.q.ReviewInfo.WithContext(ctx).
		Where(r.data.q.ReviewInfo.ReviewID.Eq(reviewReply.ReviewID)).
		First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return v1.ErrorReviewNotFound("评价%v不存在", reviewReply.ReviewID)
		}
		r.log.Errorf("data CreateReply failed, err:%v\n", err)
		return v1.ErrorInternalError("内部错误")
	}
	if review.HasReply > 0 {
		return v1.ErrorHasBeenReplied("订单：%v已被回复过", reviewReply.ReviewID)
	}

	//业务校验--防止商家水平越权
	if review.StoreID != reviewReply.StoreID {
		return v1.ErrorInvalidParam("水平越权！禁止商家%v给订单评价%v回复", reviewReply.StoreID, reviewReply.ReviewID)
	}

	//事务操作 更新评价字段 & 创建回复
	return r.data.q.Transaction(func(tx *query.Query) error {
		_, err = tx.ReviewInfo.WithContext(ctx).
			Where(tx.ReviewInfo.ReviewID.Eq(reviewReply.ReviewID)).
			Update(tx.ReviewInfo.HasReply, 1)
		if err != nil {
			r.log.Errorf("data CreateReply Transaction failed, err:%v\n", err)
			return err
		}

		if err = tx.ReviewReplyInfo.WithContext(ctx).Create(reviewReply); err != nil {
			r.log.Errorf("data CreateReply Transaction failed, err:%v\n", err)
			return err
		}
		return nil
	})
}

// CreateAppeal B端 申诉评价
func (r *reviewRepo) CreateAppeal(ctx context.Context, ra *model.ReviewAppealInfo) (*model.ReviewAppealInfo, error) {
	//必须有效的评价id
	_, err := r.data.q.ReviewInfo.WithContext(ctx).
		Where(r.data.q.ReviewInfo.ReviewID.Eq(ra.ReviewID)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, v1.ErrorReviewNotFound("评价%v不存在", ra.ReviewID)
		}
		r.log.Errorf("data CreateAppeal failed, err:%v\n", err)
		return nil, v1.ErrorInternalError("内部错误")
	}

	//必须未申诉过
	appeal, err := r.data.q.ReviewAppealInfo.WithContext(ctx).
		Where(r.data.q.ReviewAppealInfo.ReviewID.Eq(ra.ReviewID)).
		First()
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		r.log.Errorf("data CreateAppeal failed, err:%v\n", err)
		return nil, v1.ErrorInternalError("内部错误")
	}
	if appeal != nil {
		return nil, v1.ErrorHasBeenAppealed("订单%v已被商家申诉过", ra.ReviewID)
	}

	//防止商家水平越权 (理论上来说需有此业务逻辑，此处省略）

	err = r.data.q.ReviewAppealInfo.WithContext(ctx).Create(ra)
	if err != nil {
		r.log.Errorf("data CreateAppeal failed, err:%v\n", err)
		return nil, err
	}
	return ra, nil
}

// AuditAppeal O端 审核申诉
func (r *reviewRepo) AuditAppeal(ctx context.Context, ra *model.ReviewAppealInfo) error {
	//业务逻辑 （appealID判断、reviewID判断、已审核判断）
	appeal, err := r.data.q.ReviewAppealInfo.WithContext(ctx).
		Where(r.data.q.ReviewAppealInfo.AppealID.Eq(ra.AppealID)).
		First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return v1.ErrorAppealNotFound("申诉%v不存在", ra.AppealID)
		}
		r.log.Errorf("[data] AuditAppeal failed, err:%v\n", err)
		return err
	}

	if appeal.Status > 10 {
		return v1.ErrorAppealHasBeenAudit("申诉%v已被审核过", ra.AppealID)
	}

	_, err = r.data.q.ReviewInfo.WithContext(ctx).
		Where(r.data.q.ReviewInfo.ReviewID.Eq(ra.ReviewID)).
		First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return v1.ErrorReviewNotFound("评价%v不存在", ra.ReviewID)
		}
		r.log.Errorf("[data] AuditAppeal failed, err:%v\n", err)
		return err
	}

	//事务操作 审核通过则隐藏评价
	return r.data.q.Transaction(func(tx *query.Query) error {
		_, err = tx.ReviewAppealInfo.WithContext(ctx).
			Where(tx.ReviewAppealInfo.AppealID.Eq(ra.AppealID)).
			Updates(ra)
		if err != nil {
			r.log.Errorf("[data] AuditAppeal failed, err:%v\n", err)
			return err
		}

		if ra.Status == 20 {
			_, err = tx.ReviewInfo.WithContext(ctx).
				Where(tx.ReviewInfo.ReviewID.Eq(ra.ReviewID)).
				Update(tx.ReviewInfo.Status, 40)
			if err != nil {
				r.log.Errorf("[data] AuditAppeal failed, err:%v\n", err)
				return err
			}
		}
		return nil
	})
}
