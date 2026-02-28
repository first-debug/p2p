package explorer

type Explorer interface {
	StartRecive() chan []byte
	Send([]byte) error
}
