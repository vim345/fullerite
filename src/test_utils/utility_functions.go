package test_utils

import (
	l "github.com/Sirupsen/logrus"
)

func BuildLogger() *l.Entry {
	return l.WithFields(l.Fields{"testing": true})
}
