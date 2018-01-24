package strategy

const (
	SIG_NONE = iota // none
	SIG_RISE        // going up
	SIG_FALL        // going down
)

// Strategy is trading strategy
type Strategy interface {
	Name() string
	Signal() (uint8, error)
}

// Strategies return all available strategy
func Strategies() map[string]Strategy {
	ripdog := RippleDoge{}

	return map[string]Strategy{
		ripdog.Name(): &ripdog,
	}
}

// Available return all available strategy name
func Available() []string {
	m := Strategies()
	keys := make([]string, 0, len(m))

	for k := range m {
		keys = append(keys, k)
	}

	return keys
}

// Signals return all available signal
func Signals() []uint8 {
	return []uint8{
		SIG_NONE,
		SIG_RISE,
		SIG_FALL,
	}
}
