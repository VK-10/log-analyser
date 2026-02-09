package classifier

import "time"

func Retry(attempts int, fn func() (string, error)) (string, error) {
	var err error
	var label string

	for i := 0; i < attempts; i++ {
		label, err = fn()
		if err == nil {
			return label, nil
		}
		time.Sleep(time.Duration(i+1) * 100 * time.Millisecond)
	}

	return "", err
}
