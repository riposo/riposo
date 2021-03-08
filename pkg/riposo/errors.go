package riposo

// ErrCode is a specific internal error code.
type ErrCode int

// Error codes enum.
const (
	ErrCodeMissingAuthToken      ErrCode = 104
	ErrCodeInvalidAuthToken      ErrCode = 105
	ErrCodeBadJSON               ErrCode = 106
	ErrCodeInvalidParameters     ErrCode = 107
	ErrCodeMissingParameters     ErrCode = 108
	ErrCodeInvalidPostedData     ErrCode = 109
	ErrCodeInvalidResourceID     ErrCode = 110
	ErrCodeMissingResource       ErrCode = 111
	ErrCodeMissingContentLength  ErrCode = 112
	ErrCodeRequestTooLarge       ErrCode = 113
	ErrCodeModifiedMeanwhile     ErrCode = 114
	ErrCodeMethodNotAllowed      ErrCode = 115
	ErrCodeVersionNotAvailable   ErrCode = 116
	ErrCodeClientReachedCapacity ErrCode = 117
	ErrCodeForbidden             ErrCode = 121
	ErrCodeConstraintViolated    ErrCode = 122
	ErrCodeBackend               ErrCode = 201
	ErrCodeServiceDeprecated     ErrCode = 202
	ErrCodeUndefined             ErrCode = 999
)
