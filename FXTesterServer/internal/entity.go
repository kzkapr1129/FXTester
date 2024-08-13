package internal

import "time"

type UserEntity struct {
	UserId                int64
	Email                 string
	RefreshToken          string
	RefreshTokenExpiresAt time.Time
}
