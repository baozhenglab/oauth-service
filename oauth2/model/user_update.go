package model

import (
	"github.com/200lab/go-sdk/sdkcm"
	"github.com/200lab/oauth-service/common"
)

type UserUpdate struct {
	Id                   string       `json:"-" gorm:"-"`
	Username             *string      `json:"username" form:"username" gorm:"username"`
	Password             *string      `json:"password" form:"password" gorm:"password"`
	PasswordConfirmation *string      `json:"password_confirmation" form:"password_confirmation" gorm:"-"`
	Email                *string      `json:"email,omitempty" form:"email" gorm:"email"`
	PhonePrefix          *string      `json:"phone_prefix" form:"phone_prefix" bson:"phone_prefix,omitempty" gorm:"phone_prefix"`
	Phone                *string      `json:"phone" form:"phone" bson:"phone,omitempty" gorm:"phone"`
	AccountType          *AccountType `json:"account_type" form:"account_type" bson:"account_type" gorm:"account_type"`
	FBId                 *string      `json:"fb_id" form:"fb_id" bson:"fb_id" gorm:"fb_id"`
	AKId                 *string      `json:"ak_id" form:"ak_id" bson:"account_kit_id" gorm:"column:account_kit_id"`
	AppleId              *string      `json:"apple_id" form:"apple_id" bson:"apple_id" gorm:"apple_id"`
	ClientId             *string      `json:"client_id" form:"client_id" bson:"client_id"`
	Status               *int         `json:"status" form:"status" gorm:"status"`
	Salt                 *string      `json:"-" gorm:"salt"`
}

func (u *UserUpdate) Validate() error {
	if u.Email != nil && *u.Email != "" {
		if !common.IsValidEmail(*u.Email) {
			err := common.ErrEmailInvalid
			return sdkcm.ErrInvalidRequestWithMessage(err, err.Error())
		}
	}

	if u.Username != nil && *u.Username != "" {
		if *u.Username == "" {
			err := common.ErrUsernameCannotBeEmpty
			return sdkcm.ErrInvalidRequestWithMessage(err, err.Error())
		}
	}

	if u.Password != nil && *u.Password != "" {
		if u.PasswordConfirmation == nil || *u.Password != *u.PasswordConfirmation {
			err := common.ErrPassAndConfirmNotMatch
			return sdkcm.ErrInvalidRequestWithMessage(err, err.Error())
		}
	}

	return nil
}
