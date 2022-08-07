package queue

type ServiceDataResponse struct {
	Rejected  int64 `json:"rejected"`
	New       int64 `json:"new"`
	Consumers int64 `json:"consumers"`
	InWork    int64 `json:"in_work"`
}
