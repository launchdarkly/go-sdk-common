package ldcontext

const (
	// AttrNameKey is a constant for the attribute name that corresponds to the Key() method in
	// Context and Builder. This is the name of the attribute in JSON representations and flag rules.
	AttrNameKey = "key"

	// AttrNameKind is a constant for the attribute name that corresponds to the Kind() method in
	// Context and Builder. This is the name of the attribute in JSON representations and flag rules.
	AttrNameKind = "kind"

	// AttrNameName is a constant for the attribute name that corresponds to the Name() method in
	// Context and Builder. This is the name of the attribute in JSON representations and flag rules.
	AttrNameName = "name"

	// AttrNameTransient is a constant for the attribute name that corresponds to the Transient()
	// method in Context and Builder. This is the name of the attribute in JSON representations and flag rules.
	AttrNameTransient = "transient"
)

const (
	jsonPropMeta             = "_meta"
	jsonPropPrivate          = "privateAttributeNames"
	jsonPropSecondary        = "secondary"
	jsonPropOldUserAnonymous = "anonymous"
	jsonPropOldUserCustom    = "custom"
)
