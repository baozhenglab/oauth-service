package model

import "time"

type AccountType string

const (
	// account logged in from other source: FB/Google/Github
	AccTypeExternal = "external"
	// account logged with username and password
	AccTypeInternal = "internal"
	// account external but has set username and password
	AccTypeBoth = "both"
)

type Provider struct {
	Id     string
	Source string
}

type User struct {
	UserId              string      `json:"id" bson:"-" gorm:"-"`
	Username            *string     `json:"username" gorm:"username"`
	Password            string      `json:"-"`
	Salt                string      `json:"-"`
	Email               *string     `json:"email"`
	PhonePrefix         *string     `json:"phone_prefix" bson:"phone_prefix,omitempty" gorm:"phone_prefix"`
	Phone               *string     `json:"phone" bson:"phone,omitempty"`
	AccountType         AccountType `json:"account_type" bson:"account_type" gorm:"account_type"`
	FBId                *string     `json:"fb_id,omitempty" bson:"fb_id" gorm:"fb_id"`
	AKId                string      `json:"ak_id" bson:"account_kit_id" gorm:"column:account_kit_id"`
	AppleId             *string     `json:"apple_id" bson:"apple_id" gorm:"apple_id"`
	ClientId            string      `json:"client_id" bson:"client_id"`
	IsNew               bool        `json:"is_new" bson:"-" gorm:"-"`
	OtpCode             *string     `json:"otp_code" bson:"otp_code" gorm:"otp_code"`
	OtpCodeExpiredAt    *time.Time  `json:"otp_code_expired_at" bson:"otp_code_expired_at" gorm:"otp_code_expired_at"`
	HasUsernamePassword bool        `json:"has_username_password" bson:"-" gorm:"-"`
}

func (u User) GetUsername() string {
	if u.Username == nil {
		return ""
	}
	return *u.Username
}

func (u User) GetEmail() string {
	if u.Email == nil {
		return ""
	}
	return *u.Email
}

func (u User) GetUserID() string {
	return u.UserId
}

type CredentialAndPassword struct {
	Id       string  `json:"id" gorm:"id"`
	Username string  `json:"username" form:"username" gorm:"username"`
	Password *string `json:"password" form:"password" gorm:"password"`
	Salt     *string `json:"salt" gorm:"salt"`
	Email    *string `json:"email" form:"email"`
	Phone    *string `json:"phone" form:"phone" bson:"phone,omitempty"`
	ClientId string  `json:"client_id" gorm:"client_id"`
}

func (up *CredentialAndPassword) Map() map[string]interface{} {
	result := make(map[string]interface{})

	if up.Username != "" {
		result["username"] = up.Username
	}

	if up.Password != nil {
		result["password"] = up.Password
		result["salt"] = up.Salt
	}

	if up.Email != nil {
		result["email"] = up.Email
	}

	if up.Phone != nil {
		result["phone"] = up.Phone
	}

	return result
}

type UserFilter struct {
	UserId      *string `json:"id" bson:"-" gorm:"-"`
	Username    *string `json:"username" form:"username" gorm:"username"`
	Email       *string `json:"email" form:"email"`
	PhonePrefix *string `json:"phone_prefix" form:"phone_prefix" bson:"phone_prefix,omitempty" gorm:"phone_prefix"`
	Phone       *string `json:"phone" form:"phone" bson:"phone,omitempty"`
	FBId        *string `json:"fb_id" form:"fb_id" bson:"fb_id" gorm:"fb_id"`
	AKId        *string `json:"ak_id" form:"ak_id" bson:"account_kit_id" gorm:"column:account_kit_id"`
	AppleId     *string `json:"apple_id" form:"apple_id" bson:"apple_id" gorm:"apple_id"`
	OTPCode     *string `json:"otp_code" form:"otp_code" bson:"otp_code" gorm:"otp_code"`
	ClientId    *string `json:"client_id" form:"client_id" bson:"client_id"`
}

func (uf *UserFilter) Map() map[string]interface{} {
	result := make(map[string]interface{})

	if v := uf.UserId; v != nil {
		result["id"] = v
	}

	if v := uf.Username; v != nil {
		result["username"] = v
	}

	if v := uf.Email; v != nil {
		result["email"] = v
	}

	if v := uf.Phone; v != nil {
		result["phone"] = v
	}

	if v := uf.FBId; v != nil {
		result["fb_id"] = v
	}

	if v := uf.AppleId; v != nil {
		result["apple_id"] = v
	}

	if v := uf.OTPCode; v != nil {
		result["otp_code"] = v
	}

	return result
}
