package data

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"
	v1 "reviewService/api/review/v1"
	"reviewService/internal/biz"
	"reviewService/internal/data/model"
	"reviewService/internal/data/query"
	"strconv"
	"strings"
	"time"
)

var g singleflight.Group

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

// GetByOrderID  根据order id 获取评价
func (r *reviewRepo) GetByOrderID(ctx context.Context, orderID int64) (*model.ReviewInfo, error) {
	return r.data.q.WithContext(ctx).ReviewInfo.
		Where(r.data.q.ReviewInfo.OrderID.Eq(orderID)).
		First()
}

// Save C端 创建评价
func (r *reviewRepo) Save(ctx context.Context, review *model.ReviewInfo) (*model.ReviewInfo, error) {
	return review, r.data.q.ReviewInfo.WithContext(ctx).Create(review)
}

// ListReviewByStoreID C端 根据商家ID获取评价列表
func (r *reviewRepo) ListReviewByStoreID(ctx context.Context, storeID int64, offset int64, limit int64) ([]*biz.MyReviewInfo, error) {
	//利用singleflight 合并重复查询请求 避免缓存击穿
	key := fmt.Sprintf("review:%v:%v:%v", storeID, offset, limit)
	b, err := r.getDataFromSingleflight(ctx, key)
	if err != nil {
		return nil, err
	}

	//从结果中解析出想要数据
	hm := new(types.HitsMetadata)
	err = json.Unmarshal(b, hm)
	if err != nil {
		r.log.Errorf("data review ListReviewByStoreID key:%v failed, err:%v\n", key, err)
		return nil, err
	}

	reviewInfos := make([]*biz.MyReviewInfo, 0, hm.Total.Value)
	for _, hit := range hm.Hits {
		reviewInfo := new(biz.MyReviewInfo)
		//log.Debug(string(hit.Source_))
		err = json.Unmarshal(hit.Source_, reviewInfo)
		if err != nil {
			r.log.Warnf("data review ListReviewByStoreID key:%v failed, err:%v\n", key, err)
			continue
		}
		reviewInfos = append(reviewInfos, reviewInfo)
	}

	return reviewInfos, nil
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

// singleflight
func (r *reviewRepo) getDataFromSingleflight(ctx context.Context, key string) ([]byte, error) {
	v, err, _ := g.Do(key, func() (interface{}, error) {
		//先从redis缓存查询
		bs, err := r.getDataFromCache(ctx, key)

		if err != nil {
			//查询不到缓存 去ES查，然后重新存入缓存
			if errors.Is(err, redis.Nil) {
				bs, err = r.getDataFromES(ctx, key)
				if err != nil {
					return nil, err
				}
				return bs, r.setCache(ctx, key, bs)
			}
			//出错直接return 避免压力下放
			return nil, err
		}

		//查询到了缓存
		return bs, nil
	})
	if err != nil {
		return nil, err
	}
	return v.([]byte), nil
}

// redis
// 取数据
func (r *reviewRepo) getDataFromCache(ctx context.Context, key string) ([]byte, error) {
	b, err := r.data.rdb.Get(ctx, key).Bytes()
	if err != nil {
		r.log.Warnf("data review getDataFromCache key:%v failed, err:%v\n", key, err)
		return nil, err
	}
	return b, nil
}

// 存数据
func (r *reviewRepo) setCache(ctx context.Context, key string, data any) error {
	return r.data.rdb.Set(ctx, key, data, time.Minute*5).Err()
}

// ES
// 取数据
func (r *reviewRepo) getDataFromES(ctx context.Context, key string) ([]byte, error) {
	split := strings.Split(key, ":")
	if len(split) < 4 {
		return nil, errors.New("getDataFromES invalid key")
	}
	index, storeID, offsetStr, limitStr := split[0], split[1], split[2], split[3]

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		r.log.Errorf("data review getDataFromES key:%v failed, err:%v\n", key, err)
		return nil, err
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		r.log.Errorf("data review getDataFromES key:%v failed, err:%v\n", key, err)
		return nil, err
	}

	resp, err := r.data.es.Search().
		Index(index).
		From(offset).
		Size(limit).
		Query(&types.Query{
			Bool: &types.BoolQuery{
				Filter: []types.Query{
					{
						Term: map[string]types.TermQuery{
							"store_id": types.TermQuery{
								Value: storeID,
							},
						},
					},
				},
			},
		}).Do(ctx)
	if err != nil {
		r.log.Errorf("data review getDataFromES key:%v failed, err:%v\n", key, err)
		return nil, err
	}

	return json.Marshal(resp.Hits)

}
