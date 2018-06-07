package project

type persistentFile struct {
	Path string
}

type envVariable struct {
	Name      string
	Value     []byte
	Encrypted bool
}
