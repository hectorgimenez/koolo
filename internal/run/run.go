package run

type Run interface {
	MoveToStartingPoint() error
	TravelToDestination() error
	Kill() error
}
