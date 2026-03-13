package test

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
)

const NotFound = "NOT_FOUND"

type ActionLine struct {
	Action     string
	Key        string
	Value      int
	IsNotFound bool
}

func (suite *QuiteDbTestSuite) loadActionLines() ([]ActionLine, error) {
	file, err := os.Open(ActionLog)
	suite.Assertions.NoError(err)
	defer file.Close()

	var lines []ActionLine

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var actionLine ActionLine
		actionLine, err = ParseLine(scanner.Text())
		suite.Assertions.NoError(err)
		lines = append(lines, actionLine)
	}
	return lines, scanner.Err()
}

func ParseLine(line string) (ActionLine, error) {
	strArr := strings.Split(line, " ")
	action := strArr[0]
	key := strArr[1]
	var value int64
	isNotFound := strArr[2] == NotFound
	if !isNotFound {
		value, _ = strconv.ParseInt(strArr[2], 10, 64)

	}
	return ActionLine{Action: action, Key: key, Value: int(value), IsNotFound: isNotFound}, nil

}

func (suite *QuiteDbTestSuite) makeRequest(action ActionLine) (string, error) {
	if action.Action == "PUT" {
		putEndpoint := fmt.Sprintf("%s/put/%s", suite.serverUrl, action.Key)
		buffer := bytes.NewBuffer([]byte(fmt.Sprintf("%d", action.Value)))
		resp, err := http.Post(putEndpoint, "text/plain", buffer)
		defer resp.Body.Close()
		suite.Assertions.NoError(err)
		data, err := io.ReadAll(resp.Body)
		return string(data), err
	} else {
		getEndpoint := fmt.Sprintf("%s/get/%s", suite.serverUrl, action.Key)
		resp, err := http.Get(getEndpoint)
		defer resp.Body.Close()
		suite.Assertions.NoError(err)
		data, err := io.ReadAll(resp.Body)
		return string(data), err
	}
}
