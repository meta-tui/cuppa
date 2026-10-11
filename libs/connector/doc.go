// Package connector is the live connection between a running application and
// the Cuppa designer (ADR 0016): the application publishes the inputs of its
// pages and the designer sends back the events a click or key raises. Its
// code lives in the slice packages below this directory:
//
//	message  the JSON-lines protocol: what is said and how a line is read
//	address  the rule that a connection stays on this computer
//	serve    the application's side: listen, publish, receive events
//	dial     the designer's side: connect, receive states, send events
//
// It is headless and uses only the standard library.
package connector
