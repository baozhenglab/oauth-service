package usrrepo

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/baozhenglab/oauth-service/common"
	"github.com/baozhenglab/oauth-service/oauth2/model"
	"github.com/baozhenglab/oauth-service/secure"
	"github.com/baozhenglab/sdkcm"
)

const (
	minOtp             = 100000
	maxOtp             = 999999
	otpExpireInSeconds = 60
)

func (ur *userRepository) GenerateOTP(ctx context.Context, userFilter *model.UserFilter) (string, error) {
	oldUser, err := ur.storage.Find(ctx, userFilter.Map())

	if err != nil {
		return "", sdkcm.ErrWithMessage(err, common.ErrDataNotFound)
	}

	otp := fmt.Sprintf("%d", rand.New(rand.NewSource(time.Now().UnixNano())).Intn(maxOtp-minOtp)+minOtp)

	if err := ur.storage.Update(ctx,
		map[string]interface{}{"id": oldUser.UserId},
		map[string]interface{}{
			"otp_code":            otp,
			"otp_code_expired_at": time.Now().UTC().Add(time.Second * otpExpireInSeconds),
		},
	); err != nil {
		return "", sdkcm.ErrDB(err)
	}

	return otp, nil
}

func (ur *userRepository) LoginWithOTP(ctx context.Context, userFilter *model.UserFilter) (*model.User, error) {
	//if userFilter.OTPCode == nil {
	//	return nil, sdkcm.ErrWithMessage(errors.New("otp code is blank"), common.ErrDataNotFound)
	//}

	oldUser, err := ur.storage.Find(ctx, userFilter.Map())

	if err != nil {
		return nil, sdkcm.ErrWithMessage(err, common.ErrDataNotFound)
	}

	if otpExp := oldUser.OtpCodeExpiredAt; otpExp != nil {
		if otpExp.Before(time.Now().UTC()) {
			return nil, sdkcm.ErrWithMessage(err, common.ErrOTPExpired)
		}
	}

	oldUser.User.UserId = fmt.Sprintf("%d", oldUser.ID)

	_ = ur.storage.Update(
		ctx,
		map[string]interface{}{"id": oldUser.UserId},
		map[string]interface{}{
			"otp_code":            nil,
			"otp_code_expired_at": nil,
		},
	)

	return &oldUser.User, nil
}

func (ur *userRepository) LoginWithOtherCredentialAndPassword(ctx context.Context, credential *model.CredentialAndPassword) (*model.User, error) {
	password := *credential.Password
	credential.Password = nil

	oldUser, err := ur.storage.Find(ctx, credential.Map())

	if err != nil {
		return nil, sdkcm.ErrWithMessage(err, common.ErrDataNotFound)
	}

	if oldUser.Password != secure.ComputeHmac256(password, oldUser.Salt, ur.sm.GetSystemSecret()) {
		return nil, sdkcm.ErrCustom(nil, common.ErrCannotLogin)
	}

	oldUser.User.UserId = fmt.Sprintf("%d", oldUser.ID)

	return &oldUser.User, nil
}
