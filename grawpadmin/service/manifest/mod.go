package manifest

type Manifest interface {
	formatString(string, string) (string, error)
	GetManifestDirectory() string
}
