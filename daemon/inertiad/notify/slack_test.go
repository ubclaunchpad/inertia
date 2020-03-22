package notify

import "testing"

func TestSlackNotifier_IsEqual(t *testing.T) {
	type fields struct {
		hookURL string
	}
	type args struct {
		nt Notifier
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{"ok: same hook url", fields{"abcde"}, args{&SlackNotifier{"abcde"}}, true},
		{"not ok: diff hook url", fields{"robert"}, args{&SlackNotifier{"abcde"}}, false},
		{"not ok: not slack notifier", fields{"robert"}, args{nil}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &SlackNotifier{
				hookURL: tt.fields.hookURL,
			}
			if got := n.IsEqual(tt.args.nt); got != tt.want {
				t.Errorf("SlackNotifier.IsEqual() = %v, want %v", got, tt.want)
			}
		})
	}
}
