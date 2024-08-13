package db

type UserEntity struct {
	UserId       int64
	Email        string
	AccessToken  *string
	RefreshToken *string
}
