package manifest

const defaultFileMode = 0755

type Manifest interface {
	formatString(string, string) (string, error)
	GetManifestDirectory() string
}
