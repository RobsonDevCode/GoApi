package models

type ApiResponse[T any] struct {
	Data  T
	Error error
}
