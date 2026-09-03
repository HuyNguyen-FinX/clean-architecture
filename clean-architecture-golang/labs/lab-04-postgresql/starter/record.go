package starter

import "errors"

var ErrNotFound = errors.New("not found")

// AccountRecord cố ý đóng cả ba vai trò DB row, domain object và output DTO.
type AccountRecord struct {
	ID             string
	Balance        int64
	Currency       string
	OverdraftLimit int64
	Status         string
}

type Store struct {
	rows map[string]AccountRecord
}

func NewStore(rows ...AccountRecord) *Store {
	s := &Store{rows: make(map[string]AccountRecord)}
	for _, row := range rows {
		s.rows[row.ID] = row
	}
	return s
}

func (s *Store) FindByID(id string) (AccountRecord, error) {
	row, ok := s.rows[id]
	if !ok {
		return AccountRecord{}, ErrNotFound
	}
	return row, nil
}
