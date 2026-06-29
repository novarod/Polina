package hash

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
)

func Mission(missionData any) (string, error) {
	b, err := json.Marshal(missionData)
	if err != nil {
		return "", fmt.Errorf("hash: marshal missionData: %w", err)
	}
	sum := sha256.Sum256(b)
	return fmt.Sprintf("%x", sum), nil
}

func APIKey(rawKey string) string {
	sum := sha256.Sum256([]byte(rawKey))
	return fmt.Sprintf("%x", sum)
}
