package bounty

import (
	"log"
)

func CleanupAfterRewardCreation() error {
	// Some code that might fail
	if err := DeleteChore(); err != nil {
		log.Printf("Error during cleanup after reward creation: %v", err)
		return err
	}
	return nil
}

func CleanupAfterSecondStep() error {
	// Some code that might fail
	if err := DeleteChore(); err != nil {
		log.Printf("Error during cleanup after second step: %v", err)
		return err
	}
	return nil
}

func DeleteChore() error {
	// Implementation of DeleteChore
	return nil
}