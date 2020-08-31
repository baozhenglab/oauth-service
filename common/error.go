package common

// Error helpers
var (
	ErrDataNotFound                     = CustomError("ErrDataNotFound", "data not found")
	ErrOldPassNotCorrect                = CustomError("ErrOldPassNotCorrect", "old password is not correct")
	ErrPassAndConfirmNotMatch           = CustomError("ErrPassAndConfirmNotMatch", "password & password confirmation not match")
	ErrWrongClientId                    = CustomError("ErrWrongClientId", "wrong client id")
	ErrUsernameAndPasswordCannotBeEmpty = CustomError("ErrUsernameAndPasswordCannotBeEmpty", "username and password cannot be empty")
	ErrUserExisted                      = CustomError("ErrUserExisted", "user is existed")
	ErrUsernameExisted                  = CustomError("ErrUsernameExisted", "username is existed")
	ErrEmailExisted                     = CustomError("ErrEmailExisted", "email is existed")
	ErrFbIdCannotBeEmpty                = CustomError("ErrFbIdCannotBeEmpty", "Facebook id cannot be empty")
	ErrAKIdCannotBeEmpty                = CustomError("ErrAKIdCannotBeEmpty", "AccountKit id cannot be empty")
	ErrAppleIdCannotBeEmpty             = CustomError("ErrAppleIdCannotBeEmpty", "Apple id cannot be empty")
	ErrPhonePrefixCannotBeEmpty         = CustomError("ErrPhonePrefixCannotBeEmpty", "phone prefix cannot be empty")
	ErrEmailCannotBeEmpty               = CustomError("ErrEmailCannotBeEmpty", "email cannot be empty")
	ErrPhoneAndEmailCannotBeEmpty       = CustomError("ErrPhoneAndEmailCannotBeEmpty", "phone or email must be have a value")
	ErrEmailInvalid                     = CustomError("ErrEmailInvalid", "email is not valid format example@email.com")
	ErrUsernameCannotBeEmpty            = CustomError("ErrUsernameCannotBeEmpty", "username cannot be empty")
	ErrOTPExpired                       = CustomError("ErrOTPExpired", "otp expired")
	ErrCannotLogin                      = CustomError("ErrCannotLogin", "cannot login, wrong credential")
)

type customError struct {
	k string
	v string
}

func (ce *customError) Error() string {
	return ce.v
}

func (ce *customError) Key() string {
	return ce.k
}

func CustomError(k, v string) *customError {
	return &customError{k, v}
}
