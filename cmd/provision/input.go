package provisioncmd

import (
	"fmt"
	"os"

	"github.com/ubclaunchpad/inertia/cmd/core/utils/out"
)

// enterEC2CredentialsWalkthrough prints promts to stdout and reads input from
// given reader
func enterEC2CredentialsWalkthrough() (id, key string, err error) {
	out.Print(`To get your credentials:
	1. Open the IAM console (https://console.aws.amazon.com/iam/home?#home).
	2. In the navigation pane of the console, choose Users. You may have to create a user.
	3. Choose your IAM user name (not the check box).
	4. Choose the Security credentials tab and then choose Create access key.
	5. To see the new access key, choose Show. Your credentials will look something like this:

		Access key ID: AKIAIOSFODNN7EXAMPLE
		Secret access key: wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
	`)

	var response string

	out.Print("\nKey ID:       ")
	_, err = fmt.Fscanln(os.Stdin, &response)
	if err != nil {
		return
	}
	id = response

	out.Print("\nAccess Key:   ")
	_, err = fmt.Fscanln(os.Stdin, &response)
	if err != nil {
		return
	}
	key = response
	return
}
