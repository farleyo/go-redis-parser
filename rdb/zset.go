package rdb

import (
	"encoding/json"
	"fmt"
	"strconv"
)

type SortedSet struct {
	Key     KeyObject
	Len     uint64
	Entries []SortedSetEntry
}

type SortedSetEntry struct {
	Field interface{}
	Score float64
}

func (r *ParseRdb) readZSet(key KeyObject, t byte) error {
	length, _, err := r.loadLen()
	if err != nil {
		return err
	}
	sortedSet := SortedSet{Key: key, Len: length, Entries: make([]SortedSetEntry, 0, length)}
	for i := uint64(0); i < length; i++ {
		member, err := r.loadString()
		if err != nil {
			return err
		}
		var score float64
		if t == TypeZset2 {
			score, err = r.loadBinaryFloat()
		} else {
			score, err = r.loadFloat()
		}
		if err != nil {
			return err
		}
		sortedSet.Entries = append(sortedSet.Entries, SortedSetEntry{Field: ToString(member), Score: score})
	}
	r.data <- sortedSet

	return nil
}

func (r *ParseRdb) readZipListSortSet(key KeyObject) error {
	b, err := r.loadString()
	if err != nil {
		return err
	}
	buf := newInput(b)
	cardinality, err := loadZiplistLength(buf)
	if err != nil {
		return err
	}
	cardinality /= 2

	sortedSet := SortedSet{Key: key, Len: uint64(cardinality), Entries: make([]SortedSetEntry, 0, cardinality)}
	for i := int64(0); i < cardinality; i++ {
		member, err := loadZiplistEntry(buf)
		if err != nil {
			return err
		}
		scoreBytes, err := loadZiplistEntry(buf)
		if err != nil {
			return err
		}
		score, err := strconv.ParseFloat(string(scoreBytes), 64)
		if err != nil {
			return err
		}
		sortedSet.Entries = append(sortedSet.Entries, SortedSetEntry{Field: ToString(member), Score: score})
	}
	r.data <- sortedSet

	return nil
}

func (zs SortedSet) String() string {
	itemStr, _ := json.Marshal(zs.Entries)
	return fmt.Sprintf("SortedSetMetadata{Key: %s, Len: %d, Entries: %s}", ToString(zs.Key), zs.Len, ToString(itemStr))
}