package domain

import "slices"

type Status string

const (
	StatusPending   Status = "PENDING"
	StatusPickedUp  Status = "PICKED_UP"
	StatusInTransit Status = "IN_TRANSIT"
	StatusDelivered Status = "DELIVERED"
	StatusCancelled Status = "CANCELLED"
)

var validTransitions = map[Status][]Status{
	StatusPending:   {StatusPickedUp, StatusCancelled},
	StatusPickedUp:  {StatusInTransit, StatusCancelled},
	StatusInTransit: {StatusDelivered, StatusCancelled},
	StatusDelivered: {},
	StatusCancelled: {},
}

var allStatuses = []Status{
	StatusPending,
	StatusPickedUp,
	StatusInTransit,
	StatusDelivered,
	StatusCancelled,
}

func (s Status) IsTerminal() bool {
	return s == StatusDelivered || s == StatusCancelled
}

func (s Status) CanTransitionTo(next Status) bool {
	allowed, ok := validTransitions[s]
	if !ok {
		return false
	}
	return slices.Contains(allowed, next)
}

func (s Status) IsValid() bool {
	return slices.Contains(allStatuses, s)
}
