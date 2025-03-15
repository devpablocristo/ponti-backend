package support

import "errors"

func GetCredentials(username, email, password string) (string, string, error) {
	if username == "" && email == "" {
		return "", "", errors.New("username and email are empty")
	}
	if password == "" {
		return "", "", errors.New("password is empty")
	}

	var nameCredential string
	if username == "" {
		nameCredential = email
	} else {
		nameCredential = username
	}

	return nameCredential, password, nil
}
