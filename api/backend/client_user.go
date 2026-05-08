package backend

// UserGetter returns the currently authenticated user.
type UserGetter interface {
	GetCurrentUser() (User, error)
}
