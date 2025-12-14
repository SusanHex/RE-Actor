package actions

type Action interface {
	Act(string) error
}
