package main

import (
	"net/url"
	"os/exec"
)

func makeQStatArgString(qstatID string, statURLs []url.URL) []string {
	argString := []string{"-xml", "-utf8", "-R", "-P", "-" + qstatID}

	for _, v := range statURLs {
		argString = append(argString, v.String())
	}

	argString = append(argString, "-")

	return argString
}

// GetQStatOutput spuns up QStat and reads XML output.
func GetQStatOutput(info GameInfo, s SettingsMap) ([]string, error) {
	var statURLs []url.URL
	var qstatID string

	output, err := exec.Command("qstat", makeQStatArgString(qstatID, statURLs)...).Output()

	return []string{string([]byte(output))}, err
}
