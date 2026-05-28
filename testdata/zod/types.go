// Package zod provides sample types for testing the Zod serializer.
package zod

import (
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	StatusActive  Status = "active"
	StatusPending Status = "pending"
	StatusClosed  Status = "closed"
)

type Priority int

const (
	PriorityLow    Priority = 0
	PriorityMedium Priority = 1
	PriorityHigh   Priority = 2
)

// Base is embedded by Ticket to test heritage/extend.
type Base struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Ticket demonstrates a realistic struct with enums, nullable
// pointers, embedded structs, arrays, and maps.
type Ticket struct {
	Base

	Title       string            `json:"title"`
	Description *string           `json:"description,omitempty"`
	Status      Status            `json:"status"`
	Priority    Priority          `json:"priority"`
	AssigneeID  *uuid.UUID        `json:"assignee_id,omitempty"`
	Tags        []string          `json:"tags"`
	Metadata    map[string]string `json:"metadata"`
	Children    []Ticket          `json:"children"`
}

// CreateTicketRequest demonstrates a request body type.
type CreateTicketRequest struct {
	Title       string   `json:"title"`
	Description string   `json:"description,omitempty"`
	Priority    Priority `json:"priority"`
	Tags        []string `json:"tags,omitempty"`
}
