package state

type Persistence interface {
	Load() (SerializedState, error)
	Save(state SerializedState) error
}
