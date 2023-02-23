package api

import (
	"encoding/json"
	"strconv"
	"time"
)

func (self *IoK8sApimachineryPkgApisMetaV1Time) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Time(*self))
}

func (self *IoK8sApimachineryPkgApisMetaV1Time) UnmarshalJSON(b []byte) error {
	return json.Unmarshal(b, (*time.Time)(self))
}

func (self *IoK8sApimachineryPkgUtilIntstrIntOrString) MarshalJSON() ([]byte, error) {
	i, err := strconv.Atoi(string(*self))
	if err == nil {
		return json.Marshal(i)
	} else {
		return json.Marshal(string(*self))
	}
}
