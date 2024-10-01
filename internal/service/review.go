package service

import (
	"context"
	pb "reviewService/api/review/v1"
	"reviewService/internal/biz"
	"reviewService/internal/data/model"
)

type ReviewService struct {
	pb.UnimplementedReviewServer

	uc *biz.ReviewUsecase
}

// NewReviewService review服务 构造函数
func NewReviewService(uc *biz.ReviewUsecase) *ReviewService {
	return &ReviewService{uc: uc}
}

// CreateReview 创建评价
func (s *ReviewService) CreateReview(ctx context.Context, req *pb.CreateReviewRequest) (*pb.CreateReviewReply, error) {
	anonymous := 0
	if req.GetAnonymous() {
		anonymous = 1
	}

	//转换格式
	review, err := s.uc.CreateReview(ctx, &model.ReviewInfo{
		Content:      req.GetContent(),
		Score:        req.GetScore(),
		ServiceScore: req.GetServiceScore(),
		ExpressScore: req.GetExpressScore(),
		OrderID:      req.GetOrderID(),
		StoreID:      req.GetStoreID(),
		UserID:       req.GetUserID(),
		Anonymous:    int32(anonymous),
		PicInfo:      req.GetPicInfo(),
		VideoInfo:    req.GetVideoInfo(),
	})
	if err != nil {
		return nil, err
	}
	return &pb.CreateReviewReply{ReviewID: review.ReviewID}, nil
}
func (s *ReviewService) UpdateReview(ctx context.Context, req *pb.UpdateReviewRequest) (*pb.UpdateReviewReply, error) {
	return &pb.UpdateReviewReply{}, nil
}
func (s *ReviewService) DeleteReview(ctx context.Context, req *pb.DeleteReviewRequest) (*pb.DeleteReviewReply, error) {
	return &pb.DeleteReviewReply{}, nil
}
func (s *ReviewService) GetReview(ctx context.Context, req *pb.GetReviewRequest) (*pb.GetReviewReply, error) {
	return &pb.GetReviewReply{}, nil
}
func (s *ReviewService) ListReview(ctx context.Context, req *pb.ListReviewRequest) (*pb.ListReviewReply, error) {
	return &pb.ListReviewReply{}, nil
}

// ReplyReview B端 回复评价
func (s *ReviewService) ReplyReview(ctx context.Context, req *pb.ReplyReviewRequest) (*pb.ReplyReviewReply, error) {
	reviewReply := &model.ReviewReplyInfo{
		ReviewID:  req.GetReviewID(),
		StoreID:   req.GetStoreID(),
		Content:   req.GetContent(),
		PicInfo:   req.GetPicInfo(),
		VideoInfo: req.GetVideoInfo(),
	}
	err := s.uc.CreateReply(ctx, reviewReply)
	if err != nil {
		return nil, err
	}
	return &pb.ReplyReviewReply{ReplyID: reviewReply.ReplyID}, nil
}

// AppealReview B端 申诉评价
func (s *ReviewService) AppealReview(ctx context.Context, req *pb.AppealReviewRequest) (*pb.AppealReviewReply, error) {
	appeal, err := s.uc.AppealReview(ctx, &model.ReviewAppealInfo{
		ReviewID:  req.GetReviewID(),
		StoreID:   req.GetStoreID(),
		Reason:    req.GetReason(),
		Content:   req.GetContent(),
		PicInfo:   req.GetPicInfo(),
		VideoInfo: req.GetVideoInfo(),
	})
	if err != nil {
		return nil, err
	}
	return &pb.AppealReviewReply{AppealID: appeal.AppealID}, nil
}

// AuditReview O端 审核评价
func (s *ReviewService) AuditReview(ctx context.Context, req *pb.AuditReviewRequest) (*pb.AuditReviewReply, error) {
	err := s.uc.AuditReview(ctx, &model.ReviewInfo{
		ReviewID:  req.GetReviewID(),
		Status:    req.GetStatus(),
		OpReason:  req.GetOpReason(),
		OpRemarks: req.GetOpRemarks(),
		OpUser:    req.GetOpUser(),
	})
	if err != nil {
		return nil, err
	}
	return &pb.AuditReviewReply{
		ReviewID: req.ReviewID,
		Status:   req.Status,
	}, nil
}

// AuditAppeal O端 审核申诉
func (s *ReviewService) AuditAppeal(ctx context.Context, req *pb.AuditAppealRequest) (*pb.AuditAppealReply, error) {
	err := s.uc.AuditAppeal(ctx, &model.ReviewAppealInfo{
		AppealID:  req.GetAppealID(),
		ReviewID:  req.GetReviewID(),
		Status:    req.GetStatus(),
		OpRemarks: req.GetOpRemarks(),
		OpUser:    req.GetOpUser(),
	})
	if err != nil {
		return nil, err
	}
	return &pb.AuditAppealReply{}, nil
}
