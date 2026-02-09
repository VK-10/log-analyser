package classifier

import "context"

func CallLLMWithTimeout(ctx context.Context, msg string) (string, error) {
	resultCh := make(chan string, 1)
	errCh := make(chan error, 1)

	go func() {
		label, err := callLLMInternal(msg)
		if err != nil {
			errCh <- err
			return
		}
		resultCh <- label
	}()

	select {
	case label := <-resultCh:
		return label, nil
	case err := <-errCh:
		return "", err
	case <-ctx.Done():
		return "", ctx.Err()
	}
}
