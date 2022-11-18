package ldap

import "testing"

func TestGetUID(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedResult string
		expectedError  string
	}{
		{
			name:           "valid input",
			input:          "uid=avulaj,ou=users,dc=redhat,dc=com",
			expectedResult: "avulaj",
		},
		{
			name:          "no uid present",
			input:         "ou=users,dc=redhat,dc=com",
			expectedError: "no uid field found for given ldap string",
		},
		{
			name:          "malformed dn",
			input:         "uid:avulaj",
			expectedError: "error parsing dn: DN ended with incomplete type, value pair",
		},
		{
			name:          "empty input",
			expectedError: "no uid field found for given ldap string",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := getUID(tc.input)
			if tc.expectedError != "" && err.Error() != tc.expectedError {
				t.Fatalf("Did not receive the expected error.\nExpected: %v\nActual: %v", tc.expectedError, err.Error())
			}
			if result != tc.expectedResult {
				t.Fatalf("Expected %v, but got %v", tc.expectedResult, result)
			}
		})
	}
}
