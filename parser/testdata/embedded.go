package embedded

type User struct {
    ID   int
    Name string
}

type Admin struct {
    User           // embedded
    Level int
}

type Reader interface {
    Read(p []byte) (int, error)
}

type Writer interface {
    Write(p []byte) (int, error)
}

type ReadWriter interface {
    Reader         // embedded
    Writer         // embedded
}