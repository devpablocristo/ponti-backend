package pkgsuggester

func Bootstrap(opts ...Option) (*Suggester, error) {
	cfg, err := newConfig(opts...)
	if err != nil {
		return nil, err
	}
	return newSuggester(cfg), nil
}
