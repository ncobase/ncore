package ctxutil

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"testing"
	"time"
)

func TestWithAsyncContext(t *testing.T) {
	parent, parentCancel := context.WithCancel(context.Background())
	defer parentCancel()

	asyncCtx, asyncCancel := WithAsyncContext(parent, 100*time.Millisecond)
	defer asyncCancel()

	parentCancel()
	select {
	case <-asyncCtx.Done():
		if asyncCtx.Err() == context.Canceled {
			t.Fatal("async context should not be canceled directly by parent cancellation")
		}
	default:
	}

	timeoutCtx, timeoutCancel := WithAsyncContext(context.Background(), 10*time.Millisecond)
	defer timeoutCancel()
	<-timeoutCtx.Done()
	if timeoutCtx.Err() != context.DeadlineExceeded {
		t.Fatalf("expected DeadlineExceeded, got %v", timeoutCtx.Err())
	}
}

func TestWithAsyncContextDefault(t *testing.T) {
	ctx, cancel := WithAsyncContextDefault(context.Background())
	defer cancel()

	deadline, ok := ctx.Deadline()
	if !ok {
		t.Fatal("WithAsyncContextDefault() should set a deadline")
	}

	remaining := time.Until(deadline)
	if remaining < 4*time.Second || remaining > 6*time.Second {
		t.Fatalf("unexpected default timeout window: %v", remaining)
	}
}

func TestGenerateBusinessCode(t *testing.T) {
	identifier := fmt.Sprintf("T%d", time.Now().UnixNano()%100000)
	codePattern := regexp.MustCompile("^" + identifier + `\d{6}[A-Z0-9]{3}\d{4}$`)

	code1 := GenerateBusinessCode(identifier)
	code2 := GenerateBusinessCode(identifier)

	if !codePattern.MatchString(code1) {
		t.Fatalf("GenerateBusinessCode() returned invalid format: %q", code1)
	}
	if !codePattern.MatchString(code2) {
		t.Fatalf("GenerateBusinessCode() returned invalid format: %q", code2)
	}

	serial1, err := strconv.Atoi(code1[len(code1)-4:])
	if err != nil {
		t.Fatalf("failed to parse serial number from code1: %v", err)
	}
	serial2, err := strconv.Atoi(code2[len(code2)-4:])
	if err != nil {
		t.Fatalf("failed to parse serial number from code2: %v", err)
	}
	if serial2 != serial1+1 {
		t.Fatalf("expected serial increment by 1, got %d then %d", serial1, serial2)
	}
}

func TestGenerateAntifakeCode(t *testing.T) {
	code := generateAntifakeCode(8)
	if len(code) != 8 {
		t.Fatalf("generateAntifakeCode() expected length 8, got %d", len(code))
	}

	for _, r := range code {
		if !(r >= 'A' && r <= 'Z') && !(r >= '0' && r <= '9') {
			t.Fatalf("generateAntifakeCode() returned invalid character: %q", string(r))
		}
	}
}
