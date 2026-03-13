// Package chans implements helpers and operators to perform common channel patterns.
//
// Operators almost always introduce a buffer of size 1 and spawn goroutines, so please
// consider that channels that have operators on them might not follow the same semantics
// of simple channels. For example the fact that a value was sent over a channel
// might still mean that the result channel of an operator might not be ready to be received from.
//
// Almost all operators accept a done channel. That channel reports that the operation
// should be stopped. This is most commonly the channel returned by calling context.Context.Done(),
// but it may be any channel that is used to signal that the current operation is not needed
// anymore.
//
// Operators that return a channel will close the returned channel when the source is
// closed or done is closed.
//
// Returned channels don't need to be drained if done is closed.
package chans
