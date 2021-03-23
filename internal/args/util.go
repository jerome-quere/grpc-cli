package args

import "strings"

// SplitRaw creates a slice that maps arg names to their values.
// ["arg1=1", "arg2=2", "arg3"] => { {"arg1", "1"}, {"arg2", "2"}, {"arg3",""} }
func SplitRaw(rawArgs []string) [][2]string {
	keyValue := [][2]string{}
	for _, arg := range rawArgs {
		tmp := strings.SplitN(arg, "=", 2)
		if len(tmp) < 2 {
			tmp = append(tmp, "")
		}
		keyValue = append(keyValue, [2]string{tmp[0], tmp[1]})
	}
	return keyValue
}
