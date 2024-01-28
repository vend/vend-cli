package messenger

// Exit is for exiting gracefully when using panic
type Exit struct {
	Code    int
	Message error
}

func ExitWithError(err error) {
	panic(Exit{1, err})
}
