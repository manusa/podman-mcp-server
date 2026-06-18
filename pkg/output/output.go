package output

type Output interface {
	Format(data interface{}) (string, error)
}

func New(format string) Output {
	switch format {
	case "json":
		return &JSONOutput{}
	default:
		return &TextOutput{}
	}
}
