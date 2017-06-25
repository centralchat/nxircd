package interfaces

// Socket Holds the basic methods for an ircd.Socket
// this allows us to override for non external clients
// like services etc
type Socket interface {
	Write(string) (int, error) // Write to the socket
	Read() (string, error)     // Read from the socket
	Close()
}
