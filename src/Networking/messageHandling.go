/*
/((([A-Za-z]{3,9}:(?:\/\/)?)(?:[-;:&=\+\$,\w]+@)?[A-Za-z0-9.-]+|(?:www.|[-;:&=\+\$,\w]+@)[A-Za-z0-9.-]+)((?:\/[\+~%\/.\w-_]*)?\??(?:[-\+=&;%@.\w_]*)#?(?:[\w]*))?)/
*/
package Networking

import (
	"regexp"
	"strconv"
)

var reCurls *regexp.Regexp
var reAngles *regexp.Regexp
var reCommandMsg *regexp.Regexp
var reIPvPort *regexp.Regexp

func setupNetworkingRegex() {
	reCurls = regexp.MustCompile(`\\{([^\\}]+)\\}`)
	reAngles = regexp.MustCompile(`<([^>]+)>`)

	reCommandMsg = regexp.MustCompile(`\{([^\{\}]+)\}(.*)`)
	reIPvPort = regexp.MustCompile(`([^:]+):(.+)`)
}
func DifferentiateMessage(incomingMsg []byte) (string, []byte) {
	result := reCommandMsg.FindSubmatch(incomingMsg)
	return string(result[1]), result[2]
}
func GetIPFromAddress(ipAddress string) (string, int) {

	result := reIPvPort.FindStringSubmatch(ipAddress)
	if port, err := strconv.Atoi(result[2]); err != nil {
		return result[1], -1
	} else {
		return result[1], port
	}
}
func SanatizeMessage(incomingMsg []byte) []byte {
	return reAngles.ReplaceAllFunc(reCurls.ReplaceAllFunc(incomingMsg,
		func(curl []byte) []byte {
			return []byte("{{" + string(curl) + "}}")
		}),
		func(angle []byte) []byte {
			return []byte("&lt" + string(angle) + "&gt")
		})
}
