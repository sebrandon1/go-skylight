if err := rewardCreation(); err != nil {
    if cleanupErr := DeleteChore(); cleanupErr != nil {
        log.Printf("failed to cleanup chore after reward creation failed: %v", cleanupErr)
    }
    return fmt.Errorf("reward creation failed: %w", err)
}

if err := secondStep(); err != nil {
    if cleanupErr := DeleteChore(); cleanupErr != nil {
        log.Printf("failed to cleanup chore after second step failed: %v", cleanupErr)
    }
    return fmt.Errorf("second step failed: %w", err)
}