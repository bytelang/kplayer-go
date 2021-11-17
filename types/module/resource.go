package module

type Resources []Resource

type ResourceError string

func (r ResourceError) Error() string {
    return string(r)
}

const (
    ResourceNotFound ResourceError = "resource not found"
)

func (rs *Resources) GetResourceByUnique(unique string) (*Resource, int, error) {
    for key, item := range *rs {
        if item.Unique == unique {
            return &(*rs)[key], key, nil
        }
    }
    return nil, 0, ResourceNotFound
}

func (rs *Resources) RemoveResourceByUnique(unique string) (*Resource, error) {
    res, index, err := rs.GetResourceByUnique(unique)
    if res == nil {
        return nil, err
    }

    var newInputs Resources
    newInputs = append(newInputs, (*rs)[:index]...)
    newInputs = append(newInputs, (*rs)[index+1:]...)

    (*rs) = newInputs
    return res, nil
}
