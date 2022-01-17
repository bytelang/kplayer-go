package error

const (
	VersionInvalidMainError    MainError = "invalid config version"
	TokenFileNotFoundMainError MainError = "token file cannot be found"
)

type MainError string

func (me MainError) Error() string {
	return string(me)
}
