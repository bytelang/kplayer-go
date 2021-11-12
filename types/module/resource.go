package module

type Resources []Resource

func (rs *Resources) GetResourceByUnique(unique string) *Resource {
    for key, item := range *rs {
        if item.Unique == unique {
            return &(*rs)[key]
        }
    }
    return nil
}
