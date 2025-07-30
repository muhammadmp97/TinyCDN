package models

type Encoding int8

const (
	EncodingNone Encoding = iota
	EncodingGZIP
)

type File struct {
	Path         string
	Content      string
	ContentPath  string
	Type         string
	Encoding     Encoding
	Size         int
	OriginalSize int
}

type Domain struct {
	Id    int16
	Name  string
	Token string
}

var Domains = []Domain{}

func GetDomain(domainName string) (found bool, domain Domain) {
	for _, domain := range Domains {
		if domain.Name == domainName {
			return true, domain
		}
	}

	return false, Domain{}
}
