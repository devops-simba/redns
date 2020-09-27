package main

type InvalidOptions struct{}

func (this InvalidOptions) Error() string {
	return "Invalid options"
}
