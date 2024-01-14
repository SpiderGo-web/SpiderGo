package engine

import "fmt"

func parseTag(tag string, file []byte) ([]string, error) {
	tagOpen, err := findLineInFile(fmt.Sprintf("<%s>", tag), file)
	if err != nil {
		return nil, err
	}
	tagClose, err := findLineInFile(fmt.Sprintf("</%s", tag), file)
	if err != nil {
		return nil, err
	}
	insideTag, err := readFileByLines(tagOpen[0].lineNumber, tagClose[0].lineNumber, file)
	if err != nil {
		return nil, err
	}
	return insideTag, nil
}
