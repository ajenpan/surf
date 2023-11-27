package msg

import "fmt"

func (err *Error) Error() string {
	return fmt.Sprintf("code:%v, msg:%v", err.Code, err.Detail)
}
