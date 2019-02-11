package v1

// Manager defines the interface for manipulating user info in the databse
//
type Manager interface {
	Create(firstName, lastName, password, email string) (ID string, err error)
}

// manager is the implementation of Manager interface
//
type manager struct {

}

// NewManager creates an instance of Manager
func NewManager() Manager {
	return &manager{}
}
