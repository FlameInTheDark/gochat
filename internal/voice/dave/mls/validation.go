package mls

import "fmt"

func ValidateKeyPackage(raw []byte) error {
	if len(raw) == 0 {
		return fmt.Errorf("empty key package")
	}
	if len(raw) < 4 {
		return fmt.Errorf("key package too short")
	}
	return nil
}

func ValidateCommit(raw []byte) error {
	if len(raw) == 0 {
		return fmt.Errorf("empty commit")
	}
	if len(raw) < 4 {
		return fmt.Errorf("commit too short")
	}
	return nil
}

func ValidateWelcome(raw []byte) error {
	if len(raw) == 0 {
		return fmt.Errorf("empty welcome")
	}
	if len(raw) < 4 {
		return fmt.Errorf("welcome too short")
	}
	return nil
}

func ValidateProposalPayloads(payloads [][]byte) error {
	if len(payloads) == 0 {
		return fmt.Errorf("empty proposal payloads")
	}
	for _, payload := range payloads {
		if len(payload) == 0 {
			return fmt.Errorf("empty proposal payload")
		}
	}
	return nil
}
