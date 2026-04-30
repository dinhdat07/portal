package portalgrpc

import (
	"context"
	adminv1 "portal-system/gen/go/admin/v1"
	commonv1 "portal-system/gen/go/common/v1"
	"portal-system/internal/domain"
	"portal-system/internal/domain/constants"
	mappers "portal-system/internal/grpc/mapper"
	"portal-system/internal/services"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	gstatus "google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

const (
	defaultPage     = 1
	defaultPageSize = 20
	maxPageSize     = 100
	dateLayout      = "2006-01-02"
)

type AdminServer struct {
	adminv1.UnimplementedAdminServiceServer
	adminService      services.AdminService
	userService       services.UserService
	roleService       services.RoleService
	permissionService services.PermissionService
}

func NewAdminServer(
	adminService services.AdminService,
	userService services.UserService,
	roleService services.RoleService,
	permissionService services.PermissionService,
) *AdminServer {
	return &AdminServer{
		adminService:      adminService,
		userService:       userService,
		roleService:       roleService,
		permissionService: permissionService,
	}
}

func (s *AdminServer) ListUsers(ctx context.Context, req *adminv1.ListUsersRequest) (*adminv1.ListUsersResponse, error) {
	if req == nil {
		return nil, gstatus.Error(codes.InvalidArgument, "request is required")
	}

	actor, err := getActorFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	page := int(req.GetPage())
	if page == 0 {
		page = defaultPage
	}

	pageSize := int(req.GetPageSize())
	if pageSize == 0 {
		pageSize = defaultPageSize
	}

	if page < 1 || pageSize < 1 || pageSize > maxPageSize {
		return nil, gstatus.Error(codes.InvalidArgument, "invalid pagination")
	}

	filter := domain.UsersFilter{
		Page:           page,
		PageSize:       pageSize,
		IncludeDeleted: req.GetIncludeDeleted(),
	}

	if req.Username != nil {
		filter.Username = req.GetUsername()
	}
	if req.Email != nil {
		filter.Email = req.GetEmail()
	}
	if req.FullName != nil {
		filter.FullName = req.GetFullName()
	}
	if req.Dob != nil {
		dob, err := time.Parse(dateLayout, req.GetDob())
		if err != nil {
			return nil, gstatus.Error(codes.InvalidArgument, "invalid dob format, expected YYYY-MM-DD")
		}
		filter.Dob = &dob
	}
	if req.RoleCode != nil {
		roleCode := constants.RoleCode(req.GetRoleCode())
		filter.RoleCode = &roleCode
	}
	if req.Status != nil {
		status, ok := mappers.UserStatusFromPB(req.GetStatus())
		if !ok {
			return nil, gstatus.Error(codes.InvalidArgument, "invalid status")
		}
		filter.Status = status
	}

	meta := getAuditFromCtx(ctx)
	result, err := s.adminService.ListUsers(ctx, meta, actor, filter)
	if err != nil {
		return nil, mappers.MapError(err)
	}

	return mappers.ListUsersResultToPB(result), nil
}

func (s *AdminServer) CreateUser(ctx context.Context, req *adminv1.CreateUserRequest) (*commonv1.User, error) {
	if req == nil {
		return nil, gstatus.Error(codes.InvalidArgument, "request is required")
	}

	actor, err := getActorFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	meta := getAuditFromCtx(ctx)

	dob, err := time.Parse(dateLayout, req.GetDob())
	if err != nil {
		return nil, gstatus.Error(codes.InvalidArgument, "invalid dob format, expected YYYY-MM-DD")
	}

	input := domain.CreateUserInput{
		Email:     req.GetEmail(),
		Username:  req.GetUsername(),
		FirstName: req.GetFirstName(),
		LastName:  req.GetLastName(),
		RoleCode:  constants.RoleCode(req.GetRoleCode()),
		DOB:       &dob,
	}

	user, err := s.adminService.CreateUser(ctx, meta, actor, input)
	if err != nil {
		return nil, mappers.MapError(err)
	}

	return mappers.UserModelToPB(user), nil
}

func (s *AdminServer) GetUserDetail(ctx context.Context, req *adminv1.GetUserDetailRequest) (*commonv1.User, error) {
	if req == nil {
		return nil, gstatus.Error(codes.InvalidArgument, "request is required")
	}

	actor, err := getActorFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	userID, err := parseUserID(req.GetUserId())
	if err != nil {
		return nil, err
	}

	meta := getAuditFromCtx(ctx)
	user, err := s.userService.GetProfile(ctx, meta, actor, userID)
	if err != nil {
		return nil, mappers.MapError(err)
	}

	return mappers.UserModelToPB(user), nil
}

func (s *AdminServer) UpdateUser(ctx context.Context, req *adminv1.UpdateUserRequest) (*commonv1.User, error) {
	if req == nil {
		return nil, gstatus.Error(codes.InvalidArgument, "request is required")
	}

	actor, err := getActorFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	userID, err := parseUserID(req.GetUserId())
	if err != nil {
		return nil, err
	}

	input := domain.UpdateUserInput{}

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
		dob, err := time.Parse(dateLayout, req.GetDob())
		if err != nil {
			return nil, gstatus.Error(codes.InvalidArgument, "invalid dob format, expected YYYY-MM-DD")
		}
		input.DOB = &dob
	}

	meta := getAuditFromCtx(ctx)
	user, err := s.userService.UpdateProfile(ctx, meta, actor, userID, input)
	if err != nil {
		return nil, mappers.MapError(err)
	}

	return mappers.UserModelToPB(user), nil
}

func (s *AdminServer) DeleteUser(ctx context.Context, req *adminv1.DeleteUserRequest) (*commonv1.User, error) {
	if req == nil {
		return nil, gstatus.Error(codes.InvalidArgument, "request is required")
	}

	actor, err := getActorFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	userID, err := parseUserID(req.GetUserId())
	if err != nil {
		return nil, err
	}

	meta := getAuditFromCtx(ctx)
	user, err := s.adminService.DeleteUser(ctx, meta, actor, userID)
	if err != nil {
		return nil, mappers.MapError(err)
	}

	return mappers.UserModelToPB(user), nil
}

func (s *AdminServer) RestoreUser(ctx context.Context, req *adminv1.RestoreUserRequest) (*commonv1.User, error) {
	if req == nil {
		return nil, gstatus.Error(codes.InvalidArgument, "request is required")
	}

	actor, err := getActorFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	userID, err := parseUserID(req.GetUserId())
	if err != nil {
		return nil, err
	}

	meta := getAuditFromCtx(ctx)
	user, err := s.adminService.RestoreUser(ctx, meta, actor, userID)
	if err != nil {
		return nil, mappers.MapError(err)
	}

	return mappers.UserModelToPB(user), nil
}

func (s *AdminServer) UpdateUserRole(ctx context.Context, req *adminv1.UpdateUserRoleRequest) (*commonv1.User, error) {
	if req == nil {
		return nil, gstatus.Error(codes.InvalidArgument, "request is required")
	}

	actor, err := getActorFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	userID, err := parseUserID(req.GetUserId())
	if err != nil {
		return nil, err
	}

	meta := getAuditFromCtx(ctx)
	user, err := s.adminService.UpdateRole(ctx, meta, actor, userID, constants.RoleCode(req.GetRoleCode()))
	if err != nil {
		return nil, mappers.MapError(err)
	}

	return mappers.UserModelToPB(user), nil
}

func (s *AdminServer) ListRoles(ctx context.Context, _ *emptypb.Empty) (*adminv1.ListRolesResponse, error) {
	_, err := getActorFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	roles, err := s.roleService.ListRoles(ctx)
	if err != nil {
		return nil, mappers.MapError(err)
	}

	return mappers.RolesToPB(roles), nil
}

func (s *AdminServer) CreateRole(ctx context.Context, req *adminv1.CreateRoleRequest) (*commonv1.Role, error) {
	if req == nil {
		return nil, gstatus.Error(codes.InvalidArgument, "request is required")
	}

	actor, err := getActorFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	meta := getAuditFromCtx(ctx)
	role, err := s.roleService.CreateRole(ctx, meta, actor, domain.CreateRoleInput{
		Code: constants.RoleCode(req.GetCode()),
		Name: req.GetName(),
	})
	if err != nil {
		return nil, mappers.MapError(err)
	}

	return mappers.RoleModelToPB(role), nil
}

func (s *AdminServer) DeleteRole(ctx context.Context, req *adminv1.DeleteRoleRequest) (*commonv1.MessageResponse, error) {
	if req == nil {
		return nil, gstatus.Error(codes.InvalidArgument, "request is required")
	}

	actor, err := getActorFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	roleID, err := parseRoleID(req.GetRoleId())
	if err != nil {
		return nil, err
	}

	var replacementRoleID *uuid.UUID
	if req.ReplacementRoleId != nil {
		replacementID, err := parseRoleID(req.GetReplacementRoleId())
		if err != nil {
			return nil, gstatus.Error(codes.InvalidArgument, "invalid replacement_role_id")
		}
		replacementRoleID = &replacementID
	}

	meta := getAuditFromCtx(ctx)
	if err := s.roleService.DeleteRole(ctx, meta, actor, roleID, replacementRoleID); err != nil {
		return nil, mappers.MapError(err)
	}

	return &commonv1.MessageResponse{Message: "role deleted successfully"}, nil
}

func (s *AdminServer) AssignPermissionToRole(ctx context.Context, req *adminv1.AssignPermissionToRoleRequest) (*commonv1.MessageResponse, error) {
	if req == nil {
		return nil, gstatus.Error(codes.InvalidArgument, "request is required")
	}

	actor, err := getActorFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	roleID, err := parseRoleID(req.GetRoleId())
	if err != nil {
		return nil, err
	}
	permissionID, err := parsePermissionID(req.GetPermissionId())
	if err != nil {
		return nil, err
	}

	meta := getAuditFromCtx(ctx)
	if err := s.roleService.AssignPermission(ctx, meta, actor, roleID, permissionID); err != nil {
		return nil, mappers.MapError(err)
	}

	return &commonv1.MessageResponse{Message: "permission assigned successfully"}, nil
}

func (s *AdminServer) RemovePermissionFromRole(ctx context.Context, req *adminv1.RemovePermissionFromRoleRequest) (*commonv1.MessageResponse, error) {
	if req == nil {
		return nil, gstatus.Error(codes.InvalidArgument, "request is required")
	}

	actor, err := getActorFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	roleID, err := parseRoleID(req.GetRoleId())
	if err != nil {
		return nil, err
	}
	permissionID, err := parsePermissionID(req.GetPermissionId())
	if err != nil {
		return nil, err
	}

	meta := getAuditFromCtx(ctx)
	if err := s.roleService.RemovePermission(ctx, meta, actor, roleID, permissionID); err != nil {
		return nil, mappers.MapError(err)
	}

	return &commonv1.MessageResponse{Message: "permission removed successfully"}, nil
}

func (s *AdminServer) ListPermissions(ctx context.Context, _ *emptypb.Empty) (*adminv1.ListPermissionsResponse, error) {
	_, err := getActorFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	perms, err := s.permissionService.ListPermission(ctx)
	if err != nil {
		return nil, mappers.MapError(err)
	}

	return mappers.PermissionsToPB(perms), nil
}

func parseUserID(id string) (uuid.UUID, error) {
	return parseUUIDField(id, "user_id")
}

func parseRoleID(id string) (uuid.UUID, error) {
	return parseUUIDField(id, "role_id")
}

func parsePermissionID(id string) (uuid.UUID, error) {
	return parseUUIDField(id, "permission_id")
}

func parseUUIDField(id string, field string) (uuid.UUID, error) {
	if id == "" {
		return uuid.Nil, gstatus.Error(codes.InvalidArgument, field+" is required")
	}

	parsedID, err := uuid.Parse(id)
	if err != nil {
		return uuid.Nil, gstatus.Error(codes.InvalidArgument, "invalid "+field)
	}

	return parsedID, nil
}
