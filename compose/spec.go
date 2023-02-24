package compose

type File struct {
	Version  string
	Services map[string]Service
}

type Service struct {
	Image string
}
