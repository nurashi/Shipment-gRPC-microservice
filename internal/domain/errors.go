package domain

import "errors"

var (
	ErrShipmentNotFound         = errors.New("shipment not found")
	ErrInvalidStatusTransition  = errors.New("invalid status transition")
	ErrShipmentTerminated       = errors.New("shipment is in a terminal state")
	ErrDuplicateReferenceNumber = errors.New("duplicate reference number")
	ErrMissingRequiredField     = errors.New("missing required field")
	ErrInvalidFieldValue        = errors.New("invalid field value")
)
