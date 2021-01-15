package bricks

import (
	"encoding/json"
	"errors"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

var reField = regexp.MustCompile("[^0-9A-Za-z_]")
var reValue = regexp.MustCompile("['\r\n\t]")

// Update the store with new map[string]interface{}, register changes.
func Update(db *sqlx.DB, label string, data map[string]map[string]interface{}) error {
	incoming := make(map[string]interface{})
	for id, incomingRecord := range data {
		outRecord := make(map[string]interface{})
		for inKey, inValue := range incomingRecord {
			outKey := strings.ToLower(reField.ReplaceAllString(inKey, ""))
			outValue := reValue.ReplaceAllString(inValue.(string), "")
			outRecord[outKey] = outValue
		}
		incoming[id] = outRecord
	}

	existing, err := Read(db, label, Live)
	if err != nil {
		return err
	}

	changes := Compare(existing, incoming)
	if len(changes) > 0 {
		mappedChanges := make(map[string]interface{})
		for _, change := range changes {
			id, mapped := change.Map()
			mappedChanges[id] = mapped
		}
		if err := Write(db, label, Changes, mappedChanges); err != nil {
			return err
		}

		if err := Write(db, label, Live, incoming); err != nil {
			return err
		}
	}

	return nil
}

// Restore the `archive` table to a given time.
func Restore(db *sqlx.DB, label string, target time.Time) (map[string]interface{}, error) {
	data, err := Read(db, label, Live)
	if err != nil {
		return nil, err
	}

	rawChanges, err := Read(db, label, Changes)
	if err != nil {
		return nil, err
	}
	var changes []Change
	for k, v := range rawChanges {
		record := v.(map[string]interface{})

		keychain := record["keychain"].(string)
		timestamp, err := time.Parse(time.RFC3339Nano, record["timestamp"].(string))
		if err != nil {
			return nil, err
		}
		operation, err := strconv.Atoi(record["operation"].(string))
		if err != nil {
			return nil, err
		}

		old := record["old"].(string)
		new := record["new"].(string)

		change := Change{
			ID:        k,
			Keychain:  keychain,
			Timestamp: timestamp,
			Operation: Operation(operation),
			Old:       old,
			New:       new,
		}
		changes = append(changes, change)
	}
	sort.Sort(ByTimestamp(changes))

	for _, change := range changes {
		if change.Timestamp.Before(target) {
			break
		}

		keychain := strings.Split(change.Keychain, "@")
		if len(keychain) == 1 {
			switch change.Operation {
			case Addition:
				delete(data, keychain[0])
			case Deletion:
				raw, _ := change.Old.(string)

				var record map[string]interface{}
				json.Unmarshal([]byte(raw), &record)

				data[keychain[0]] = record
			}
		} else if len(keychain) == 2 {
			_, isMap := data[keychain[0]].(map[string]interface{})
			if !isMap {
				return nil, errors.New("record wasn't a map")
			}

			switch change.Operation {
			case Addition:
				delete(data[keychain[0]].(map[string]interface{}), keychain[1])
			case Modification:
				data[keychain[0]].(map[string]interface{})[keychain[1]] = change.Old
			case Deletion:
				data[keychain[0]].(map[string]interface{})[keychain[1]] = change.Old
			}
		}
	}

	return data, nil
}
