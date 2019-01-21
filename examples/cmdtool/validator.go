package main

import "log"

//type blankValidator struct{}
//
//func (blankValidator) Validate(_ string, _ []byte) error        { return nil }
//func (blankValidator) Select(_ string, _ [][]byte) (int, error) { return 0, nil }

// NullValidator is a validator that does no valiadtion
type NullValidator struct{}

// Validate always returns success
func (nv NullValidator) Validate(key string, value []byte) error {
	log.Printf("NullValidator Validate: %s - %s", key, string(value))
	return nil
}

// Select always selects the first record
func (nv NullValidator) Select(key string, values [][]byte) (int, error) {
	strs := make([]string, len(values))
	for i := 0; i < len(values); i++ {
		strs[i] = string(values[i])
	}
	log.Printf("NullValidator Select: %s - %v", key, strs)

	return 0, nil
}
