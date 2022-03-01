package sharedtest

// These functions return standardized test inputs for unmarshaling JSON user data in the old user schema.

type UnmarshalingTestParams struct { //nolint:revive
	Name string
	Data []byte
}

func MakeOldUserUnmarshalingTestParams() []UnmarshalingTestParams { //nolint:revive
	return []UnmarshalingTestParams{
		{"old user with key only", MakeOldUserWithKeyOnlyJSON()},
		{"old user with few attrs", MakeOldUserWithFewAttributesJSON()},
		{"old user with all attrs", MakeOldUserWithAllAttributesJSON()},
	}
}

func MakeOldUserWithKeyOnlyJSON() []byte { //nolint:revive
	return []byte(`{"key":"user-key"}`)
}

func MakeOldUserWithFewAttributesJSON() []byte { //nolint:revive
	return []byte(`{"key":"user-key","name":"Name","email":"test@example.com","custom":{"attr":"value"}}`)
}

func MakeOldUserWithAllAttributesJSON() []byte { //nolint:revive
	return []byte(`{"key":"user-key","secondary":"secondary-value","name":"Name","ip":"ip-value","country":"us",` +
		`"email":"test@example.com","firstName":"First","lastName":"Last","avatar":"avatar-value","anonymous":true,` +
		`"custom":{"attr":"value"}}`)
}
