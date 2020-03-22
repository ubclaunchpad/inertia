package notify

import "testing"

func TestNotifiers_Exists(t *testing.T) {
	type args struct {
		nt Notifier
	}
	tests := []struct {
		name string
		n    Notifiers
		args args
		want bool
	}{
		{"ok: exists", Notifiers{&SlackNotifier{"abcde"}}, args{&SlackNotifier{"abcde"}}, true},
		{"not ok: no notifiers", Notifiers{}, args{&SlackNotifier{"abcde"}}, false},
		{"not ok: doesnt exist", Notifiers{&SlackNotifier{"robert"}}, args{&SlackNotifier{"abcde"}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.n.Exists(tt.args.nt); got != tt.want {
				t.Errorf("Notifiers.Exists() = %v, want %v", got, tt.want)
			}
		})
	}
}
