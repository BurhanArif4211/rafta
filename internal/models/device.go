package models

import (
	"github.com/google/uuid"
	"time"
)

type Device struct {
	ID        string
	LocalIP   string
	Hostname  string
	LastSeen  time.Time
	CreatedAt time.Time
}

func NewDevice(ip, hostname string) *Device {
	now := time.Now()
	return &Device{
		ID:        uuid.New().String(),
		LocalIP:   ip,
		Hostname:  hostname,
		LastSeen:  now,
		CreatedAt: now,
	}
}
