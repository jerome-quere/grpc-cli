package args

import "fmt"

type DuplicateArgError struct {
	ArgName string
}

func (e *DuplicateArgError) Error() string {
	return fmt.Sprintf("duplicate arg %s", e.ArgName)
}

type UnmarshalArgError struct {
	ArgName  string
	ArgValue string
	Err      error
}

func (e *UnmarshalArgError) Error() string {
	return fmt.Sprintf("unmarshal error for arg %s with value %s: %s", e.ArgName, e.ArgValue, e.Err)
}

type CannotParseBoolError struct {
	Value string
}

func (e *CannotParseBoolError) Error() string {
	return fmt.Sprintf("%s is not a valid boolean value", e.Value)
}
