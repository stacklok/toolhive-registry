package registry

type OfficialRegistry struct {
	loader *Loader
}

// NewOfficialRegistry creates a new instance of the official registry
func NewOfficialRegistry(loader *Loader) *OfficialRegistry {
	return &OfficialRegistry{
		loader: loader,
	}
}

func (*OfficialRegistry) WriteJSON(path string) error {

	return nil
}
