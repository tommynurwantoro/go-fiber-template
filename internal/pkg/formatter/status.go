package formatter

type Status string

var (
	Success              Status = "success"
	CacheError           Status = "APP01"
	DatabaseError        Status = "APP02"
	InvalidRequest       Status = "APP03"
	DataNotFound         Status = "APP04"
	InternalServerError  Status = "APP05"
	DataConflict         Status = "APP06"
	Unauthorized         Status = "APP07"
	ExternalServiceError Status = "APP08"
	UnprocessableEntity  Status = "APP09"
	TooManyRequest       Status = "APP10"
)

func (s Status) String() string {
	return string(s)
}
