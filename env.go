package filterbold

import "os"

func GetOrDefault(key, fallback string) string {
	if v, ex := os.LookupEnv(key); ex {
		return v
	}

	return fallback
}
