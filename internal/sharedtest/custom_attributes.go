package sharedtest

import (
	"fmt"

	"gopkg.in/launchdarkly/go-sdk-common.v3/ldvalue"
)

const (
	SmallNumberOfCustomAttributes = 2  //nolint:revive
	LargeNumberOfCustomAttributes = 20 //nolint:revive
)

type NameAndLDValue struct { //nolint:revive
	Name  string
	Value ldvalue.Value
}

func MakeCustomAttributeNamesAndValues(count int) []NameAndLDValue { //nolint:revive
	ret := make([]NameAndLDValue, 0, count)
	for i := 1; i <= count; i++ {
		ret = append(ret, NameAndLDValue{fmt.Sprintf("attr%d", i), ldvalue.String(fmt.Sprintf("value%d", i))})
	}
	return ret
}
