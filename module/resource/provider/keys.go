package provider

type ResourceError string

func (r ResourceError) Error() string {
	return string(r)
}

const (
	ModuleName = "resource"
)

const (
	CannotRemoveCurrentResource ResourceError = "can not remove playing resource"
)
