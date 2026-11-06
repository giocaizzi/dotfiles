package catalog

// ParseArenaPage reads one datasets-server /rows page of the LMArena
// leaderboard dataset for board and returns its rows and the total row count.
func ParseArenaPage(board string, data []byte) (rows []ArenaRow, total int, err error) {
	_, _ = board, data
	return nil, 0, nil // TODO(unit 1)
}
