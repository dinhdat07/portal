package grpcserver

import (
	"context"
	commonv1 "portal-system/gen/go/common/v1"
	userv1 "portal-system/gen/go/user/v1"
	mapper "portal-system/internal/handler/grpcserver/mapper"
	"portal-system/internal/service"
	"time"

	"google.golang.org/grpc/codes"
	gstatus "google.golang.org/grpc/status"
)

type UserServer struct {
	userv1.UnimplementedUserServiceServer
	userService             service.UserService
	userNotificationService service.UserNotificationService
}

func NewUserServer(userService service.UserService, userNotificationService service.UserNotificationService) *UserServer {
	return &UserServer{
		userService:             userService,
		userNotificationService: userNotificationService,
	}
}

func (s *UserServer) GetMyProfile(ctx context.Context, req *userv1.GetMyProfileRequest) (*commonv1.User, error) {
	actor, err := getActorFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	meta := getAuditFromCtx(ctx)

	user, err := s.userService.GetProfile(ctx, meta, actor, actor.ID)
	if err != nil {
		return nil, mapper.MapError(err)
	}

	return mapper.UserModelToPB(user), nil
}

func (s *UserServer) UpdateMyProfile(ctx context.Context, req *userv1.UpdateMyProfileRequest) (*commonv1.User, error) {
	if req == nil {
		return nil, gstatus.Error(codes.InvalidArgument, "request is required")
	}

	actor, err := getActorFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	meta := getAuditFromCtx(ctx)

	input := service.UpdateUserInput{}

	if req.FirstName != nil {
		v := req.GetFirstName()
		input.FirstName = &v
	}
	if req.LastName != nil {
		v := req.GetLastName()
		input.LastName = &v
	}
	if req.Username != nil {
		v := req.GetUsername()
		input.Username = &v
	}
	if req.Dob != nil {
		dob, err := time.Parse("2006-01-02", req.GetDob())
		if err != nil {
			return nil, gstatus.Error(codes.InvalidArgument, "invalid dob format, expected YYYY-MM-DD")
		}
		input.DOB = &dob
	}

	user, err := s.userService.UpdateProfile(ctx, meta, actor, actor.ID, input)
	if err != nil {
		return nil, mapper.MapError(err)
	}

	return mapper.UserModelToPB(user), nil
}

func (s *UserServer) ChangeMyPassword(ctx context.Context, req *userv1.ChangeMyPasswordRequest) (*commonv1.MessageResponse, error) {
	if req == nil {
		return nil, gstatus.Error(codes.InvalidArgument, "request is required")
	}
	actor, err := getActorFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	meta := getAuditFromCtx(ctx)

	if err = s.userService.ChangePassword(ctx, meta, actor,
		req.GetCurrentPassword(),
		req.GetNewPassword(),
		req.GetConfirmNewPassword()); err != nil {
		return nil, mapper.MapError(err)
	}

	return &commonv1.MessageResponse{
		Message: "password changed successfully",
	}, nil
}

func (s *UserServer) ListMyNotifications(ctx context.Context, req *userv1.ListMyNotificationsRequest) (*userv1.ListMyNotificationsResponse, error) {
	if req == nil {
		return nil, gstatus.Error(codes.InvalidArgument, "request is required")
	}

	actor, err := getActorFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	page := int(req.GetPage())
	if page == 0 {
		page = 1
	}

	pageSize := int(req.GetPageSize())
	if pageSize == 0 {
		pageSize = 20
	}

	filter := service.UserNotificationListFilter{
		Page:       page,
		PageSize:   pageSize,
		UnreadOnly: req.GetUnreadOnly(),
	}

	result, err := s.userNotificationService.ListNotifications(ctx, actor.ID, filter)
	if err != nil {
		return nil, mapper.MapError(err)
	}

	return &userv1.ListMyNotificationsResponse{
		Data: mapper.ListUserNotificationsResultToPB(result),
		Meta: &commonv1.PaginationMeta{
			Page:     int32(result.Page),
			PageSize: int32(result.PageSize),
			Total:    result.Total,
		},
	}, nil
}

func (s *UserServer) GetMyNotificationDetail(ctx context.Context, req *userv1.GetMyNotificationDetailRequest) (*commonv1.UserNotification, error) {
	if req == nil {
		return nil, gstatus.Error(codes.InvalidArgument, "request is required")
	}

	actor, err := getActorFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	notificationID, err := parseUUIDField(req.GetNotificationId(), "notification_id")
	if err != nil {
		return nil, err
	}

	notification, err := s.userNotificationService.GetNotification(ctx, notificationID, actor.ID)
	if err != nil {
		return nil, mapper.MapError(err)
	}

	return mapper.UserNotificationModelToPB(notification), nil
}

func (s *UserServer) MarkNotificationAsRead(ctx context.Context, req *userv1.MarkNotificationAsReadRequest) (*commonv1.MessageResponse, error) {
	if req == nil {
		return nil, gstatus.Error(codes.InvalidArgument, "request is required")
	}

	actor, err := getActorFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	notificationID, err := parseUUIDField(req.GetNotificationId(), "notification_id")
	if err != nil {
		return nil, err
	}

	if err := s.userNotificationService.MarkAsRead(ctx, notificationID, actor.ID); err != nil {
		return nil, mapper.MapError(err)
	}

	return &commonv1.MessageResponse{Message: "notification marked as read"}, nil
}

func (s *UserServer) MarkAllNotificationsAsRead(ctx context.Context, req *userv1.MarkAllNotificationsAsReadRequest) (*commonv1.MessageResponse, error) {
	actor, err := getActorFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	if err := s.userNotificationService.MarkAllAsRead(ctx, actor.ID); err != nil {
		return nil, mapper.MapError(err)
	}

	return &commonv1.MessageResponse{Message: "all notifications marked as read"}, nil
}

func (s *UserServer) GetUnreadNotificationCount(ctx context.Context, req *userv1.GetUnreadNotificationCountRequest) (*userv1.GetUnreadNotificationCountResponse, error) {
	actor, err := getActorFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	count, err := s.userNotificationService.GetUnreadCount(ctx, actor.ID)
	if err != nil {
		return nil, mapper.MapError(err)
	}

	return &userv1.GetUnreadNotificationCountResponse{Count: count}, nil
}

func (s *UserServer) RegisterFCMToken(ctx context.Context, req *userv1.RegisterFCMTokenRequest) (*commonv1.MessageResponse, error) {
	if req == nil {
		return nil, gstatus.Error(codes.InvalidArgument, "request is required")
	}

	actor, err := getActorFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	var deviceName *string
	if req.DeviceName != nil {
		v := req.GetDeviceName()
		deviceName = &v
	}

	if err := s.userService.RegisterFCMToken(ctx, actor, req.GetFcmToken(), deviceName); err != nil {
		return nil, mapper.MapError(err)
	}

	return &commonv1.MessageResponse{Message: "fcm token registered"}, nil
}

func (s *UserServer) GenerateTelegramLinkToken(ctx context.Context, req *userv1.GenerateTelegramLinkTokenRequest) (*userv1.GenerateTelegramLinkTokenResponse, error) {
	if req == nil {
		return nil, gstatus.Error(codes.InvalidArgument, "request is required")
	}

	actor, err := getActorFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	token, link, expiresIn, err := s.userService.GenerateTelegramLinkToken(ctx, actor)
	if err != nil {
		return nil, mapper.MapError(err)
	}

	return &userv1.GenerateTelegramLinkTokenResponse{
		Token:            token,
		Link:             link,
		ExpiresInSeconds: int32(expiresIn),
	}, nil
}
