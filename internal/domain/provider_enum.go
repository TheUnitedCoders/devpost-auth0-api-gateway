// Code generated by go-enum DO NOT EDIT.
// Version:
// Revision:
// Build Date:
// Built By:

package domain

import (
	"errors"
	"fmt"
)

const (
	// RateLimitDescriptionByIp is a RateLimitDescriptionBy of type Ip.
	RateLimitDescriptionByIp RateLimitDescriptionBy = iota
	// RateLimitDescriptionBySubjectId is a RateLimitDescriptionBy of type Subject_id.
	RateLimitDescriptionBySubjectId
)

var ErrInvalidRateLimitDescriptionBy = errors.New("not a valid RateLimitDescriptionBy")

const _RateLimitDescriptionByName = "ipsubject_id"

var _RateLimitDescriptionByMap = map[RateLimitDescriptionBy]string{
	RateLimitDescriptionByIp:        _RateLimitDescriptionByName[0:2],
	RateLimitDescriptionBySubjectId: _RateLimitDescriptionByName[2:12],
}

// String implements the Stringer interface.
func (x RateLimitDescriptionBy) String() string {
	if str, ok := _RateLimitDescriptionByMap[x]; ok {
		return str
	}
	return fmt.Sprintf("RateLimitDescriptionBy(%d)", x)
}

// IsValid provides a quick way to determine if the typed value is
// part of the allowed enumerated values
func (x RateLimitDescriptionBy) IsValid() bool {
	_, ok := _RateLimitDescriptionByMap[x]
	return ok
}

var _RateLimitDescriptionByValue = map[string]RateLimitDescriptionBy{
	_RateLimitDescriptionByName[0:2]:  RateLimitDescriptionByIp,
	_RateLimitDescriptionByName[2:12]: RateLimitDescriptionBySubjectId,
}

// ParseRateLimitDescriptionBy attempts to convert a string to a RateLimitDescriptionBy.
func ParseRateLimitDescriptionBy(name string) (RateLimitDescriptionBy, error) {
	if x, ok := _RateLimitDescriptionByValue[name]; ok {
		return x, nil
	}
	return RateLimitDescriptionBy(0), fmt.Errorf("%s is %w", name, ErrInvalidRateLimitDescriptionBy)
}