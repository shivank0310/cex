package service

import (
	"context"
	"strings"

	"github.com/shivank0310/cex.git/user-service/internal/apperrors"
	"github.com/shivank0310/cex.git/user-service/internal/model"
	"github.com/shivank0310/cex.git/user-service/internal/repository"
)

type UserService struct {
	repo repository.UserStore
}

func NewUserService(repo repository.UserStore) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) Create(ctx context.Context, email, username string) (model.User, error) {
	email = strings.TrimSpace(email)
	if email == "" || !strings.Contains(email, "@") {
		return model.User{}, apperrors.New(apperrors.CodeInvalidRequest, "valid email required")
	}
	if username != "" && len(username) < 3 {
		return model.User{}, apperrors.New(apperrors.CodeInvalidRequest, "username must be at least 3 characters")
	}

	user, err := s.repo.Create(ctx, email, username)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			return model.User{}, apperrors.New(apperrors.CodeConflict, err.Error())
		}
		return model.User{}, apperrors.Wrap(apperrors.CodeInternal, "create user failed", err)
	}
	return user, nil
}

func (s *UserService) GetByID(ctx context.Context, id string) (model.User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return model.User{}, apperrors.New(apperrors.CodeNotFound, "user not found")
		}
		return model.User{}, apperrors.Wrap(apperrors.CodeInternal, "get user failed", err)
	}
	return user, nil
}

func (s *UserService) GetByEmail(ctx context.Context, email string) (model.User, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return model.User{}, apperrors.New(apperrors.CodeNotFound, "user not found")
		}
		return model.User{}, apperrors.Wrap(apperrors.CodeInternal, "get user failed", err)
	}
	return user, nil
}

func (s *UserService) Update(ctx context.Context, id string, username *string, status *string, kycStatus *string) (model.User, error) {
	var statusVal *model.Status
	var kycVal *model.KYCStatus

	if status != nil {
		st := model.Status(strings.ToUpper(strings.TrimSpace(*status)))
		if st != model.StatusActive && st != model.StatusSuspended {
			return model.User{}, apperrors.New(apperrors.CodeInvalidRequest, "status must be ACTIVE or SUSPENDED")
		}
		statusVal = &st
	}
	if kycStatus != nil {
		kyc := model.KYCStatus(strings.ToUpper(strings.TrimSpace(*kycStatus)))
		switch kyc {
		case model.KYCNone, model.KYCPending, model.KYCApproved, model.KYCRejected:
			kycVal = &kyc
		default:
			return model.User{}, apperrors.New(apperrors.CodeInvalidRequest, "invalid kyc_status")
		}
	}
	if username != nil && strings.TrimSpace(*username) != "" && len(strings.TrimSpace(*username)) < 3 {
		return model.User{}, apperrors.New(apperrors.CodeInvalidRequest, "username must be at least 3 characters")
	}

	user, err := s.repo.Update(ctx, id, username, statusVal, kycVal)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return model.User{}, apperrors.New(apperrors.CodeNotFound, "user not found")
		}
		if strings.Contains(err.Error(), "already exists") {
			return model.User{}, apperrors.New(apperrors.CodeConflict, err.Error())
		}
		return model.User{}, apperrors.Wrap(apperrors.CodeInternal, "update user failed", err)
	}
	return user, nil
}
