package sectioninvite

import "time"

type Invitation struct {
	ID        int        `json:"id"`
	ExpiresAt time.Time  `json:"expiresAt"`
	MaxUses   int        `json:"maxUses"`
	Uses      int        `json:"uses"`
	RevokedAt *time.Time `json:"revokedAt"`
	Code      string     `json:"code,omitempty"`
}

func (i *Invitation) normalizeTimes() {
	i.ExpiresAt = i.ExpiresAt.UTC()
	if i.RevokedAt != nil {
		revokedAt := i.RevokedAt.UTC()
		i.RevokedAt = &revokedAt
	}
}
