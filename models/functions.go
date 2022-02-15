package models

func (User) TableName() string {
	return "users"
}

func (Word) TableName() string {
	return "words"
}

func (Definition) TableName() string {
	return "definitions"
}

func (VerificationCode) TableName() string {
	return "verification_codes"
}
