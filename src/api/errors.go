package api

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Error codes & messages
const (
	INTERNAL_SERVER_ERROR                  = "VTS-000"
	INVALID_REQUEST                        = "VTS-001"
	USERNAME_ALREADY_TAKEN                 = "VTS-002"
	USER_DOES_NOT_EXIST                    = "VTS-003"
	USER_NAME_IS_SAME_AS_CURRENT_ONE       = "VTS-004"
	USER_LAST_NAME_IS_SAME_AS_CURRENT_ONE  = "VTS-005"
	USER_FIRST_NAME_IS_SAME_AS_CURRENT_ONE = "VTS-006"
	PLAN_NOT_FOUND                         = "VTS-004"
	PLAN_TIME_CANNOT_BE_IN_THE_PAST        = "VTS-005"
	PLAN_STATUS_INVALID                    = "VTS-005"
	PLAN_CLASH                             = "VTS-006"
)
