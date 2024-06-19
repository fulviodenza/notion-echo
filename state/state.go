package state

type IUserState interface {
	Get(userID int) string
	Set(userID int, msg string)
	Delete(userID int)
}

type UserState struct {
	states map[int]string
}

func New() IUserState {
	return &UserState{
		states: make(map[int]string),
	}
}

func (u *UserState) Get(userID int) string {
	s, ok := u.states[userID]
	if !ok {
		return ""
	}
	return s
}

func (u *UserState) Set(userID int, msg string) {
	u.states[userID] = msg
}

func (u *UserState) Delete(userID int) {
	delete(u.states, userID)
}
