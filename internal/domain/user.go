package domain

type User struct {
	GUID     string    `json:"guid" bson:"_id,omitempty"`
	Username string    `json:"username" bson:"username,omitempty"`
	Sessions []Session `json:"sessions" bson:"sessions, omitempty"`
}

type UserWithoutSession struct {
	GUID     string `json:"guid" bson:"_id,omitempty"`
	Username string `json:"username" bson:"username,omitempty"`
}
