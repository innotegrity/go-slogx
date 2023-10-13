package handler_test

import (
	"fmt"
	"log/slog"

	"go.innotegrity.dev/errorx"
)

type User struct {
	Username  string    `json:"username"`
	Password  string    `json:"password"`
	Addresses []Address `json:"addresses"`
}

type Address struct {
	Street     string `json:"street"`
	City       string `json:"city"`
	PostalCode string `json:"postal_code"`
	Country    string `json:"country"`
}

func (u User) LogValue() slog.Value {
	addrAttr := []any{}
	for i, addr := range u.Addresses {
		addrAttr = append(addrAttr, slog.Group(
			fmt.Sprintf("%03d", i),
			slog.String("street", addr.Street),
			slog.String("city", addr.City),
			slog.String("postal_code", addr.PostalCode),
			slog.String("country", addr.Country),
		))
	}
	return slog.GroupValue(
		slog.String("username", u.Username),
		slog.String("password", "********"),
		slog.Group("addresses", addrAttr...),
	)
}

const (
	ErrTestCode = 1751
)

type ErrTest struct {
	Err       error
	Value1    string
	Value2    int
	NestedErr []errorx.Error
}

// InternalError returns the internal standard error object if there is one or nil if none is set.
func (e *ErrTest) InternalError() error {
	return e.Err
}

// Error returns the string version of the error.
func (e *ErrTest) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("an error has occurred: %s", e.Err.Error())
	}
	return "an error has occurred"
}

// Code returns the corresponding error code.
func (e *ErrTest) Code() int {
	return ErrTestCode
}

func (e *ErrTest) Attrs() map[string]any {
	return map[string]any{
		"value1": e.Value1,
		"value2": e.Value2,
	}
}

func (e *ErrTest) NestedErrors() []errorx.Error {
	return e.NestedErr
}
