package msgs

type UserEventPayload struct {
	Action string `json:"action"`
	UserID uint   `json:"user_id"`
}
