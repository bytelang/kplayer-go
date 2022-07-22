package error

const (
	AuthTokenNotExists AuthError = "auth token not exists"
	AuthTokenInvalid   AuthError = "authentication failed"
)

type AuthError string

func (me AuthError) Error() string {
	return string(me)
}
