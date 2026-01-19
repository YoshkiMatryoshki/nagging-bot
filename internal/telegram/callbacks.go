package telegram

import (
	"fmt"
	"strconv"
	"strings"
)

type OccurrenceAction string

const (
	OccurrenceActionDone   OccurrenceAction = "done"
	OccurrenceActionIgnore OccurrenceAction = "ignore"
)

const occurrencePrefix = "occ"

// BuildOccurrenceCallback creates callback data for an occurrence action.
func BuildOccurrenceCallback(id int64, action OccurrenceAction) string {
	return fmt.Sprintf("%s:%d:%s", occurrencePrefix, id, action)
}

// ParseOccurrenceCallback parses callback data into action and occurrence ID.
func ParseOccurrenceCallback(data string) (action OccurrenceAction, occID int64, err error) {
	parts := strings.Split(data, ":")
	if len(parts) != 3 || parts[0] != occurrencePrefix {
		return "", 0, fmt.Errorf("unexpected format")
	}
	id, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return "", 0, err
	}
	return OccurrenceAction(parts[2]), id, nil
}
