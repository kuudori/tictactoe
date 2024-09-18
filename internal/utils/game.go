package utils

func CheckWin(board []string) string {
	// Check rows
	for i := 0; i < 9; i += 3 {
		if board[i] != "" && board[i] == board[i+1] && board[i] == board[i+2] {
			return board[i]
		}
	}

	// Check columns
	for i := 0; i < 3; i++ {
		if board[i] != "" && board[i] == board[i+3] && board[i] == board[i+6] {
			return board[i]
		}
	}

	// Check diagonals
	if board[0] != "" && board[0] == board[4] && board[0] == board[8] {
		return board[0]
	}
	if board[2] != "" && board[2] == board[4] && board[2] == board[6] {
		return board[2]
	}

	return ""
}

func IsBoardFull(board []string) bool {
	i := 0
	for _, v := range board {
		if v != "" {
			i++
		}
	}
	if i == 9 {
		return true
	}
	return false
}
