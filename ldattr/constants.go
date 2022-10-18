package ldattr

const (
	// KeyAttr is a constant for the attribute name that corresponds to the Key() method in
	// [ldcontext.Context] and [ldcontext.Builder]. This name is used in JSON representations and flag
	// rules, and can be passed to [ldcontext.Context.GetValue] or [ldcontext.Builder.SetValue].
	// representations and flag rules.
	KeyAttr = "key"

	// KindAttr is a constant for the attribute name that corresponds to the Kind() method in
	// [ldcontext.Context] and [ldcontext.Builder]. This name is used in JSON representations and flag
	// rules, and can be passed to [ldcontext.Context.GetValue] or [ldcontext.Builder.SetValue].
	// representations and flag rules.
	KindAttr = "kind"

	// NameAttr is a constant for the attribute name that corresponds to the Name() method in
	// [ldcontext.Context] and [ldcontext.Builder]. This name is used in JSON representations and flag
	// rules, and can be passed to [ldcontext.Context.GetValue] or [ldcontext.Builder.SetValue].
	// representations and flag rules.
	NameAttr = "name"

	// AnonymousAttr is a constant for the attribute name that corresponds to the Anonymous() method
	// in [ldcontext.Context] and [ldcontext.Builder]. This name is used in JSON representations and flag
	// rules, and can be passed to [ldcontext.Context.GetValue] or [ldcontext.Builder.SetValue].
	// representations and flag rules.
	AnonymousAttr = "anonymous"
)
