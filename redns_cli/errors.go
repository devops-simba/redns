package main

// This type represent invalid argument error
type InvalidArgs struct{}

func (this InvalidArgs) Error() string { return "Invalid Arguments" }

// This type represent out of range error
type OutOfRange struct{}

func (this OutOfRange) Error() string { return "Value is out of range" }
